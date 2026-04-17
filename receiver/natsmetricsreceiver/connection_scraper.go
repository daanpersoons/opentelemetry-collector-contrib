// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver"

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

type connectionScraper struct {
	cfg      Config
	settings receiver.Settings
	client   *http.Client
	mb       *metadata.MetricsBuilder
}

func (s *connectionScraper) start(ctx context.Context, host component.Host) error {
	client, err := s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	s.client = client
	s.mb = metadata.NewMetricsBuilder(s.cfg.MetricsBuilderConfig, s.settings)
	return nil
}

func (s *connectionScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	connz, err := s.fetchConnz(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	s.mb.RecordNatsConnectionCountDataPoint(now, connz.NumConnections)

	return s.mb.Emit(), nil
}

func (s *connectionScraper) fetchConnz(ctx context.Context) (*ConnzResponse, error) {
	reqURL := s.cfg.ClientConfig.Endpoint + "/connz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /connz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from /connz", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read /connz response body: %w", err)
	}

	var connz ConnzResponse
	if err := json.Unmarshal(body, &connz); err != nil {
		return nil, fmt.Errorf("failed to unmarshal /connz response: %w", err)
	}

	return &connz, nil
}

func createConnectionScraper(_ context.Context, cfg Config, settings receiver.Settings) (scraper.Metrics, error) {
	s := &connectionScraper{
		cfg:      cfg,
		settings: settings,
	}
	return scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
	)
}
