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

package internal_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/confmap/confmaptest"

	"github.com/open-telemetry/opentelemetry-collector-contrib/processor/promqlprocessor/internal"
)

func TestLoadConfig(t *testing.T) {
	t.Parallel()

	cm, err := confmaptest.LoadConf(filepath.Join("../testdata", "config", "config.yaml"))
	require.NoError(t, err)

	tests := []struct {
		id       component.ID
		expected component.Config
		validate bool
	}{
		{
			id:       component.NewIDWithName(internal.TypeStr, ""),
			expected: internal.CreateDefaultConfig(),
			validate: false,
		},
		{
			id: component.NewIDWithName(internal.TypeStr, "2"),
			expected: &internal.Config{
				TSDB: internal.TSDBConfig{
					StorageDir:               "/data",
					RetentionDurationSeconds: 0,
					MaxSamples:               999999,
					QueryTimeout:             "30ms",
					Metrics: []string{
						"a",
						"b",
						"kube_.*",
					},
				},
				LoadMetrics: internal.LoadConfig{
					MetricNames: make([]string, 0),
				},
				Queries: []internal.QueryConfig{
					{
						QueryString: `a{pod: "otel"}`,
						MetricName:  "pod_a",
					},
					{
						QueryString: `b{namespace: "otel"}`,
						MetricName:  "namespace_b",
					},
				},
			},
			validate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.id.String(), func(t *testing.T) {
			cfg := internal.CreateDefaultConfig()

			sub, err := cm.Sub(tt.id.String())
			require.NoError(t, err)
			require.NoError(t, component.UnmarshalConfig(sub, cfg))

			if tt.validate {
				assert.NoError(t, component.ValidateConfig(cfg))
			} else {
				assert.Error(t, component.ValidateConfig(cfg))
			}
			assert.Equal(t, tt.expected, cfg)
		})
	}
}
