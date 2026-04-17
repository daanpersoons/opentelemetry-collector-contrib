// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver"

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

type createNatsScraper func(context.Context, Config, receiver.Settings) (scraper.Metrics, error)

var allScrapers = map[string]createNatsScraper{
	"server":      createServerScraper,
	"connections": createConnectionScraper,
	"jetstream":   createJetStreamScraper,
	"accounts":    createAccountScraper,
}

var newMetricsReceiver = func(
	ctx context.Context,
	config Config,
	params receiver.Settings,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	scraperControllerOptions := make([]scraperhelper.ControllerOption, 0, len(config.Scrapers))
	for _, key := range config.Scrapers {
		if factory, ok := allScrapers[key]; ok {
			s, err := factory(ctx, config, params)
			if err != nil {
				return nil, err
			}
			scraperControllerOptions = append(scraperControllerOptions, scraperhelper.AddMetricsScraper(metadata.Type, s))
			continue
		}
		return nil, fmt.Errorf("no scraper found for key: %s", key)
	}

	return scraperhelper.NewMetricsController(
		&config.ControllerConfig,
		params,
		consumer,
		scraperControllerOptions...,
	)
}
