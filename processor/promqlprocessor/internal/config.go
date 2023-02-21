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
	TSDB    TSDBConfig    `mapstructure:"tsdb"`
	Queries []QueryConfig `mapstructure:"queries"`
}

type TSDBConfig struct {
	StorageDir   string   `mapstructure:"storage_dir"`
	MaxSamples   int      `mapstructure:"max_samples"`
	QueryTimeout string   `mapstructure:"query_timeout"`
	Metrics      []string `mapstructure:"metrics"`
}

type QueryConfig struct {
	QueryString      string `mapstructure:"query_string"`
	OutputMetricName string `mapstructure:"output_metric_name"`
}

func (cfg *Config) Validate() error {
	if len(cfg.Queries) == 0 || len(cfg.TSDB.Metrics) == 0 {
		return fmt.Errorf("config: empty queries or empty metrics")
	}
	return nil
}

func CreateDefaultConfig() component.Config {
	return &Config{
		TSDB: TSDBConfig{
			StorageDir:   "/tmp",
			MaxSamples:   10000000,
			QueryTimeout: "1s",
			Metrics:      make([]string, 0),
		},
		Queries: make([]QueryConfig, 0),
	}
}
