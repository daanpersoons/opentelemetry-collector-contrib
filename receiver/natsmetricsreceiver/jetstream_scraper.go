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

type jetStreamScraper struct {
	cfg      Config
	settings receiver.Settings
	client   *http.Client
	mb       *metadata.MetricsBuilder
}

func (s *jetStreamScraper) start(ctx context.Context, host component.Host) error {
	client, err := s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	s.client = client
	s.mb = metadata.NewMetricsBuilder(s.cfg.MetricsBuilderConfig, s.settings)
	return nil
}

func (s *jetStreamScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	jsz, err := s.fetchJsz(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	s.mb.RecordNatsJetstreamStreamsDataPoint(now, jsz.Streams)
	s.mb.RecordNatsJetstreamConsumersDataPoint(now, jsz.Consumers)
	s.mb.RecordNatsJetstreamMessagesDataPoint(now, jsz.Messages)
	s.mb.RecordNatsJetstreamBytesDataPoint(now, jsz.Bytes)
	s.mb.RecordNatsJetstreamAPICallsDataPoint(now, jsz.API.Total)
	s.mb.RecordNatsJetstreamAPIErrorsDataPoint(now, jsz.API.Errors)
	s.mb.RecordNatsJetstreamMemoryReservedDataPoint(now, jsz.ReservedMemory)
	s.mb.RecordNatsJetstreamStorageReservedDataPoint(now, jsz.ReservedStore)

	return s.mb.Emit(), nil
}

func (s *jetStreamScraper) fetchJsz(ctx context.Context) (*JszResponse, error) {
	reqURL := s.cfg.ClientConfig.Endpoint + "/jsz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /jsz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from /jsz", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read /jsz response body: %w", err)
	}

	var jsz JszResponse
	if err := json.Unmarshal(body, &jsz); err != nil {
		return nil, fmt.Errorf("failed to unmarshal /jsz response: %w", err)
	}

	return &jsz, nil
}

func createJetStreamScraper(_ context.Context, cfg Config, settings receiver.Settings) (scraper.Metrics, error) {
	s := &jetStreamScraper{
		cfg:      cfg,
		settings: settings,
	}
	return scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
	)
}
