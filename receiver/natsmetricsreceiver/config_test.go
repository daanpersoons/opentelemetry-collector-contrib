// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/confmap/confmaptest"
)

func TestLoadConfig(t *testing.T) {
	cm, err := confmaptest.LoadConf(filepath.Join("testdata", "config.yaml"))
	require.NoError(t, err)

	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)

	sub, err := cm.Sub("natsmetrics")
	require.NoError(t, err)
	require.NoError(t, sub.Unmarshal(cfg))

	assert.Equal(t, "http://localhost:8222", cfg.Endpoint)
	assert.Equal(t, []string{"server", "connections", "jetstream", "accounts"}, cfg.Scrapers)
	assert.Equal(t, 30*time.Second, cfg.CollectionInterval)
}

func TestValidateConfig(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		cfg := &Config{}
		cfg.Endpoint = "http://localhost:8222"
		assert.NoError(t, cfg.Validate())
	})
}
