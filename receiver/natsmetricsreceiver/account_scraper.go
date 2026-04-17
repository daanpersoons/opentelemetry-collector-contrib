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
	"go.opentelemetry.io/collector/scraper/scrapererror"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver/internal/metadata"
)

type accountScraper struct {
	cfg      Config
	settings receiver.Settings
	client   *http.Client
	mb       *metadata.MetricsBuilder
}

func (s *accountScraper) start(ctx context.Context, host component.Host) error {
	client, err := s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	s.client = client
	s.mb = metadata.NewMetricsBuilder(s.cfg.MetricsBuilderConfig, s.settings)
	return nil
}

func (s *accountScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	accstatz, err := s.fetchAccStatz(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())
	scrapeErrors := scrapererror.ScrapeErrors{}

	for _, acc := range accstatz.AccountStats {
		if acc.Acc == "" {
			scrapeErrors.AddPartial(1, fmt.Errorf("account stat entry has empty account ID, skipping"))
			continue
		}
		s.mb.RecordNatsAccountConnectionsDataPoint(now, acc.Conns, acc.Acc)
		s.mb.RecordNatsAccountSubscriptionsDataPoint(now, acc.Subs, acc.Acc)
		s.mb.RecordNatsAccountLeafNodesDataPoint(now, acc.LeafNodes, acc.Acc)
		s.mb.RecordNatsAccountMessagesSentDataPoint(now, acc.Sent.Msgs, acc.Acc)
		s.mb.RecordNatsAccountMessagesReceivedDataPoint(now, acc.Received.Msgs, acc.Acc)
		s.mb.RecordNatsAccountBytesSentDataPoint(now, acc.Sent.Bytes, acc.Acc)
		s.mb.RecordNatsAccountBytesReceivedDataPoint(now, acc.Received.Bytes, acc.Acc)
		s.mb.RecordNatsAccountSlowConsumersDataPoint(now, acc.SlowConsumers, acc.Acc)
	}

	return s.mb.Emit(), scrapeErrors.Combine()
}

func (s *accountScraper) fetchAccStatz(ctx context.Context) (*AccStatzResponse, error) {
	reqURL := s.cfg.ClientConfig.Endpoint + "/accstatz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /accstatz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from /accstatz", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read /accstatz response body: %w", err)
	}

	var accstatz AccStatzResponse
	if err := json.Unmarshal(body, &accstatz); err != nil {
		return nil, fmt.Errorf("failed to unmarshal /accstatz response: %w", err)
	}

	return &accstatz, nil
}

func createAccountScraper(_ context.Context, cfg Config, settings receiver.Settings) (scraper.Metrics, error) {
	s := &accountScraper{
		cfg:      cfg,
		settings: settings,
	}
	return scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
	)
}
