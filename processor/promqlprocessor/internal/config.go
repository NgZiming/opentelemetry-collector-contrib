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

	"go.opentelemetry.io/collector/component"
)

const (
	TypeStr = "promqlprocessor"
)

type Config struct {
	TSDB        TSDBConfig    `mapstructure:"tsdb"`
	LoadMetrics LoadConfig    `mapstructure:"load_metrics"`
	Queries     []QueryConfig `mapstructure:"queries"`

	PersistenceAcrossScrapes bool `mapstructure:"persistence_across_scrapes"`
	HonorTimestamps          bool `mapstructure:"honor_timestamps"`

	Backfilling BackfillingConfig `mapstructure:"backfilling"`
}

type TSDBConfig struct {
	StorageDir               string `mapstructure:"storage_dir"`
	RetentionDurationSeconds int64  `mapstructure:"retention_duration_seconds"`
	MaxSamples               int    `mapstructure:"max_samples"`
	Timeout                  string `mapstructure:"timeout"`

	QueryTimeout string   `mapstructure:"query_timeout"`
	Metrics      []string `mapstructure:"metrics"`
}

type LoadConfig struct {
	MetricNames []string `mapstructure:"metric_names"`
}

type QueryConfig struct {
	QueryString string `mapstructure:"query_string"`
	MetricName  string `mapstructure:"metric_name"`
}

type BackfillingConfig struct {
	SetSliceTimestampToLastestMetric bool `mapstructure:"set_slice_timestamp_to_lastest_metric"`
}

func (cfg *Config) Validate() error {
	if len(cfg.Queries) == 0 || len(cfg.TSDB.Metrics)+len(cfg.LoadMetrics.MetricNames) == 0 {
		return fmt.Errorf("config: empty queries or empty metrics")
	}

	if cfg.PersistenceAcrossScrapes || cfg.HonorTimestamps {
		fmt.Printf("PersistenceAcrossScrapes/HonorTimestamps will be ignore")
	}

	if cfg.TSDB.RetentionDurationSeconds > 0 {
		fmt.Printf("RetentionDurationSeconds will be ignore")
	}

	return nil
}

func CreateDefaultConfig() component.Config {
	return &Config{
		TSDB: TSDBConfig{
			StorageDir:               "/tmp",
			RetentionDurationSeconds: 0,
			MaxSamples:               10000000,
			Timeout:                  "",
			QueryTimeout:             "1s",
			Metrics:                  make([]string, 0),
		},
		LoadMetrics: LoadConfig{
			MetricNames: make([]string, 0),
		},
		Queries: make([]QueryConfig, 0),

		PersistenceAcrossScrapes: false,
		HonorTimestamps:          false,

		Backfilling: BackfillingConfig{
			SetSliceTimestampToLastestMetric: false,
		},
	}
}
