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
	"context"
	"os"
	"time"

	"github.com/prometheus/prometheus/promql"
	"github.com/prometheus/prometheus/storage"
	"github.com/prometheus/prometheus/tsdb"
)

type PromqlProcessorStorage struct {
	db  *tsdb.DB
	dir string
	qe  *promql.Engine
}

func NewPromqlProcessorStorage(dir string, maxSamples int, timeout string) (*PromqlProcessorStorage, error) {
	err := os.MkdirAll(dir, os.ModePerm)

	if err != nil {
		return nil, err
	}

	opts := tsdb.DefaultOptions()
	opts.MinBlockDuration = int64(24 * time.Hour / time.Millisecond)
	opts.MaxBlockDuration = int64(24 * time.Hour / time.Millisecond)
	opts.RetentionDuration = int64(time.Hour / time.Millisecond)
	db, err := tsdb.Open(dir, nil, nil, opts, tsdb.NewDBStats())

	if err != nil {
		return nil, err
	}

	promqlTimeout, _ := time.ParseDuration(timeout)
	qe := promql.NewEngine(promql.EngineOpts{
		Logger:                   nil,
		Reg:                      nil,
		MaxSamples:               maxSamples,
		Timeout:                  promqlTimeout,
		NoStepSubqueryIntervalFn: func(int64) int64 { return DurationMilliseconds(1 * time.Minute) },
		EnableAtModifier:         true,
		EnableNegativeOffset:     true,
	})

	return &PromqlProcessorStorage{db: db, dir: dir, qe: qe}, nil
}

func (pps PromqlProcessorStorage) Close() error {
	if err := pps.db.Close(); err != nil {
		return err
	}
	return os.RemoveAll(pps.dir)
}

func DurationMilliseconds(d time.Duration) int64 {
	return int64(d / (time.Millisecond / time.Nanosecond))
}

func (pps PromqlProcessorStorage) NewInstantQuery(query string, t time.Time) (promql.Query, error) {
	return pps.qe.NewInstantQuery(pps.db, nil, query, t)
}

func (pps PromqlProcessorStorage) Appender(ctx context.Context) storage.Appender {
	return pps.db.Appender(ctx)
}
