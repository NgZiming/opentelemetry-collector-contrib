// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package promqlprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/promqlprocessor"

import (
	"context"
	"path/filepath"
	"regexp"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/processor"
	"go.uber.org/multierr"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/promqlprocessor/internal"
)

type PromqlProcessor struct {
	id     string
	logger *zap.Logger

	dir string

	re *regexp.Regexp

	queries []internal.QueryConfig

	maxSamples int
	timeout    string

	c internal.Converter
}

func newPromqlProcessor(config *internal.Config, set *processor.CreateSettings) (*PromqlProcessor, error) {
	processorID := set.ID.String()
	dir := config.TSDB.StorageDir + "/" + processorID

	var exp string
	for idx, val := range config.TSDB.Metrics {
		if idx == 1 {
			exp = "(^" + val + "$)"
		} else {
			exp = exp + "|(^" + val + "$)"
		}
	}

	re, err := regexp.Compile(exp)
	if err != nil {
		return nil, err
	}

	pp := PromqlProcessor{
		id:     processorID,
		logger: set.Logger,

		dir: dir,

		re:      re,
		queries: config.Queries,

		maxSamples: config.TSDB.MaxSamples,
		timeout:    config.TSDB.QueryTimeout,

		c: internal.Converter{},
	}

	return &pp, nil
}

func (pp *PromqlProcessor) Start(context.Context, component.Host) error {
	return nil
}

func (pp *PromqlProcessor) processMetrics(ctx context.Context, pm pmetric.Metrics) (pmetric.Metrics, error) {
	resourceMetricsSlice := pm.ResourceMetrics()
	resourceMetricsSliceLen := resourceMetricsSlice.Len()

	t := time.Now()
	var rtnErr error

	for idx := 0; idx < resourceMetricsSliceLen; idx++ {
		rm := resourceMetricsSlice.At(idx)
		err := pp.processSlice(ctx, rm, t)
		if err != nil {
			pp.logger.Error("Process Resource Metrics Slice Error", zap.Error(err))
			rtnErr = multierr.Append(rtnErr, err)
		}
	}

	return pm, rtnErr
}

func (pp *PromqlProcessor) processSlice(ctx context.Context, rm pmetric.ResourceMetrics, t time.Time) error {
	var storage *internal.PromqlProcessorStorage

	{
		var err error
		storage, err = internal.NewPromqlProcessorStorage(
			filepath.Join(pp.dir, internal.TypeStr),
			pp.maxSamples,
			pp.timeout)

		if err != nil {
			pp.logger.Error("PromqlProcessor: New Local Storage Failed", zap.Error(err))
			return err
		}

		defer func() {
			err := storage.Close()
			if err != nil {
				pp.logger.Error("PromqlProcessor: Close Local Storage Failed", zap.Error(err))
			}
		}()
	}

	app := storage.Appender(ctx)
	timestampMilliseconds := t.UnixNano() / int64(time.Millisecond/time.Nanosecond)

	for i := 0; i < rm.ScopeMetrics().Len(); i++ {
		sm := rm.ScopeMetrics().At(i)
		for j := 0; j < sm.Metrics().Len(); j++ {
			m := sm.Metrics().At(j)
			if pp.re.Match([]byte(m.Name())) {
				promMetrics, err := pp.c.ExtractMetrics(m)
				if err != nil {
					pp.logger.Error("PromqlProcessor: Convert Metrics To Prometheus Format Failed", zap.Error(err))
					return err
				}

				for _, pm := range promMetrics {
					_, err = app.Append(0, pm.Labels, timestampMilliseconds, pm.Value)

					if err != nil {
						pp.logger.Error(
							"PromqlProcessor: Append Metrics to Local Storage Failed",
							zap.Error(err),
							zap.Any("metrics", pm),
							zap.Time("time", t))
						return err
					}
				}
			}
		}
	}

	err := app.Commit()
	if err != nil {
		pp.logger.Error("PromqlProcessor: Commit Metrics to Local Storage Failed", zap.Error(err))
		return err
	}

	for _, query := range pp.queries {
		e, err := storage.NewInstantQuery(query.QueryString, t)

		if err != nil {
			pp.logger.Error("PromqlProcessor: New Promql Query Failed", zap.Error(err), zap.String("query", query.QueryString))
			return err
		}

		pr := e.Exec(ctx)
		e.Close()
		if pr.Err != nil {
			pp.logger.Error("PromqlProcessor: Promql Query Failed", zap.Error(err), zap.String("query", query.QueryString))
			return pr.Err
		}

		rtnMetric := rm.ScopeMetrics().AppendEmpty()
		err = pp.c.IntoResourceMetric(pr.Value, query.OutputMetricName, rtnMetric)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pp *PromqlProcessor) Shutdown(context.Context) error {
	return nil
}
