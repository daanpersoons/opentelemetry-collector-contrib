// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer/consumertest"
	"go.opentelemetry.io/collector/receiver/receivertest"
)

var natsType = component.MustNewType("natsmetrics")

func TestNewFactory(t *testing.T) {
	factory := NewFactory()
	assert.NotNil(t, factory)
	assert.Equal(t, natsType, factory.Type())
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg)

	natsConfig := cfg.(*Config)
	assert.Equal(t, defaultEndpoint, natsConfig.Endpoint)
	assert.Equal(t, []string{"server"}, natsConfig.Scrapers)
}

func TestCreateMetricsReceiver(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	settings := receivertest.NewNopSettings(natsType)
	consumer := consumertest.NewNop()

	receiver, err := factory.CreateMetrics(context.Background(), settings, cfg, consumer)
	require.NoError(t, err)
	assert.NotNil(t, receiver)
}

func TestCreateMetricsReceiverInvalidScraper(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig().(*Config)
	cfg.Scrapers = []string{"invalid_scraper"}

	settings := receivertest.NewNopSettings(natsType)
	consumer := consumertest.NewNop()

	_, err := factory.CreateMetrics(context.Background(), settings, cfg, consumer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no scraper found for key: invalid_scraper")
}
