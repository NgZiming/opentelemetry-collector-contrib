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

package internal // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/promqlprocessor/internal"

import (
	"fmt"
	"time"

	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/promql/parser"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"

	prometheustranslator "github.com/open-telemetry/opentelemetry-collector-contrib/pkg/translator/prometheus"
)

type Converter struct{}

type PrometheusMetrics struct {
	Labels labels.Labels
	Value  float64
}

func (c Converter) ExtractMetrics(metric pmetric.Metric) ([]PrometheusMetrics, error) {
	switch metric.Type() {
	case pmetric.MetricTypeGauge:
		return c.convertGauge(metric)
	case pmetric.MetricTypeSum:
		return c.convertSum(metric)
	default:
		return nil, fmt.Errorf("unsupported types")
	}
}

func (c Converter) convertGauge(metric pmetric.Metric) ([]PrometheusMetrics, error) {
	return c.convertNumberSlice(prometheustranslator.BuildPromCompliantName(metric, ""), metric.Gauge().DataPoints())
}

func (c Converter) convertSum(metric pmetric.Metric) ([]PrometheusMetrics, error) {
	return c.convertNumberSlice(prometheustranslator.BuildPromCompliantName(metric, ""), metric.Sum().DataPoints())
}

func (c Converter) convertNumberSlice(name string, dataPoints pmetric.NumberDataPointSlice) ([]PrometheusMetrics, error) {
	l := make([]labels.Label, dataPoints.At(0).Attributes().Len()+1)
	l = append(l, labels.Label{Name: model.MetricNameLabel, Value: name})

	dataPoints.At(0).Attributes().Range(func(k string, v pcommon.Value) bool {
		l = append(l, labels.Label{
			Name:  prometheustranslator.NormalizeLabel(k),
			Value: v.AsString(),
		})
		return true
	})

	promMetrics := make([]PrometheusMetrics, dataPoints.Len())
	for idx := 0; idx < dataPoints.Len(); idx++ {
		var v float64

		switch dataPoints.At(idx).ValueType() {
		case pmetric.NumberDataPointValueTypeInt:
			v = float64(dataPoints.At(idx).IntValue())
		case pmetric.NumberDataPointValueTypeDouble:
			v = dataPoints.At(idx).DoubleValue()
		}

		promMetrics = append(promMetrics, PrometheusMetrics{
			Labels: l,
			Value:  v,
		})
	}

	return promMetrics, nil
}

func (c Converter) IntoResourceMetric(value parser.Value, name string, sm pmetric.ScopeMetrics) error {
	metric := sm.Metrics().AppendEmpty()
	metric.SetName(name)
	gauge := metric.SetEmptyGauge()

	switch value.Type() {
	case parser.ValueTypeVector:
		vector := value.(promql.Vector)
		for _, m := range vector {
			dp := gauge.DataPoints().AppendEmpty()
			dp.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(m.T)))
			dp.SetDoubleValue(m.V)
			for _, l := range m.Metric {
				if l.Name != model.MetricNameLabel {
					dp.Attributes().PutStr(l.Name, l.Value)
				}
			}
		}
	case parser.ValueTypeMatrix:
		matrix := value.(promql.Matrix)
		for _, series := range matrix {
			for _, m := range series.Points {
				dp := gauge.DataPoints().AppendEmpty()
				dp.SetTimestamp(pcommon.NewTimestampFromTime(time.UnixMilli(m.T)))
				dp.SetDoubleValue(m.V)
				for _, l := range series.Metric {
					if l.Name != model.MetricNameLabel {
						dp.Attributes().PutStr(l.Name, l.Value)
					}
				}
			}
		}
	case parser.ValueTypeNone, parser.ValueTypeScalar, parser.ValueTypeString:
		return fmt.Errorf("metrics type %s not impl", value.Type())
	default:
		return fmt.Errorf("metrics type unknown not impl")
	}

	return nil
}
