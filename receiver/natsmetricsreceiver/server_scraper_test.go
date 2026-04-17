// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/receiver/receivertest"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

func TestServerScraper(t *testing.T) {
	varz := VarzResponse{
		ServerID:         "test-server",
		Version:          "2.10.0",
		Connections:      5,
		TotalConnections: 100,
		Subscriptions:    20,
		InMsgs:           1000,
		OutMsgs:          900,
		InBytes:          10000,
		OutBytes:         9000,
		SlowConsumers:    0,
		Mem:              10485760,
		CPU:              0.5,
		Routes:           2,
		LeafNodes:        1,
		MaxConnections:   64000,
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/varz", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(varz))
	}))
	defer ts.Close()

	clientCfg := confighttp.NewDefaultClientConfig()
	clientCfg.Endpoint = ts.URL
	cfg := Config{
		ControllerConfig:     scraperhelper.NewDefaultControllerConfig(),
		ClientConfig:         clientCfg,
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}

	settings := receivertest.NewNopSettings(natsType)
	s, err := createServerScraper(context.Background(), cfg, settings)
	require.NoError(t, err)

	require.NoError(t, s.Start(context.Background(), componenttest.NewNopHost()))

	metrics, err := s.ScrapeMetrics(context.Background())
	require.NoError(t, err)
	assert.Equal(t, 13, metrics.DataPointCount())
}

func TestServerScraperFetchError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	clientCfg := confighttp.NewDefaultClientConfig()
	clientCfg.Endpoint = ts.URL
	cfg := Config{
		ControllerConfig:     scraperhelper.NewDefaultControllerConfig(),
		ClientConfig:         clientCfg,
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}

	settings := receivertest.NewNopSettings(natsType)
	s, err := createServerScraper(context.Background(), cfg, settings)
	require.NoError(t, err)

	require.NoError(t, s.Start(context.Background(), componenttest.NewNopHost()))

	_, err = s.ScrapeMetrics(context.Background())
	require.Error(t, err)
}
