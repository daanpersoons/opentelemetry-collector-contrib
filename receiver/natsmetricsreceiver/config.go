// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver"

import (
	"errors"
	"net/url"

	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

var errInvalidEndpoint = errors.New("'endpoint' must be in the form of <scheme>://<hostname>:<port>")

// Config defines the configuration for the NATS Metrics receiver.
type Config struct {
	scraperhelper.ControllerConfig `mapstructure:",squash"`
	confighttp.ClientConfig        `mapstructure:",squash"`

	// Scrapers defines which metric groups to collect.
	// Valid values: "server", "connections", "jetstream", "accounts".
	Scrapers []string `mapstructure:"scrapers"`

	// MetricsBuilderConfig allows customizing scraped metrics/attributes representation.
	metadata.MetricsBuilderConfig `mapstructure:",squash"`
}

// Validate validates the configuration fields.
func (cfg *Config) Validate() error {
	if _, err := url.Parse(cfg.Endpoint); err != nil {
		return errInvalidEndpoint
	}
	return nil
}
