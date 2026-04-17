// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver"

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

const (
	defaultEndpoint           = "http://localhost:8222"
	defaultCollectionInterval = 10 * time.Second
)

// NewFactory creates a new NATS Metrics receiver factory.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability))
}

func createDefaultConfig() component.Config {
	cfg := scraperhelper.NewDefaultControllerConfig()
	cfg.CollectionInterval = defaultCollectionInterval
	clientConfig := confighttp.NewDefaultClientConfig()
	clientConfig.Endpoint = defaultEndpoint
	return &Config{
		ControllerConfig:     cfg,
		ClientConfig:         clientConfig,
		Scrapers:             []string{"server"},
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}
}

func createMetricsReceiver(
	ctx context.Context,
	params receiver.Settings,
	cfg component.Config,
	nextConsumer consumer.Metrics,
) (receiver.Metrics, error) {
	c := cfg.(*Config)
	return newMetricsReceiver(ctx, *c, params, nextConsumer)
}
