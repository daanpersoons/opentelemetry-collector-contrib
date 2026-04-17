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

type serverScraper struct {
	cfg      Config
	settings receiver.Settings
	client   *http.Client
	mb       *metadata.MetricsBuilder
}

func (s *serverScraper) start(ctx context.Context, host component.Host) error {
	client, err := s.cfg.ClientConfig.ToClient(ctx, host.GetExtensions(), s.settings.TelemetrySettings)
	if err != nil {
		return fmt.Errorf("failed to create HTTP client: %w", err)
	}
	s.client = client
	s.mb = metadata.NewMetricsBuilder(s.cfg.MetricsBuilderConfig, s.settings)
	return nil
}

func (s *serverScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	varz, err := s.fetchVarz(ctx)
	if err != nil {
		return pmetric.Metrics{}, err
	}

	now := pcommon.NewTimestampFromTime(time.Now())

	s.mb.RecordNatsServerConnectionsDataPoint(now, varz.Connections)
	s.mb.RecordNatsServerConnectionsTotalDataPoint(now, varz.TotalConnections)
	s.mb.RecordNatsServerSubscriptionsDataPoint(now, varz.Subscriptions)
	s.mb.RecordNatsServerMessagesReceivedDataPoint(now, varz.InMsgs)
	s.mb.RecordNatsServerMessagesSentDataPoint(now, varz.OutMsgs)
	s.mb.RecordNatsServerBytesReceivedDataPoint(now, varz.InBytes)
	s.mb.RecordNatsServerBytesSentDataPoint(now, varz.OutBytes)
	s.mb.RecordNatsServerMemoryDataPoint(now, varz.Mem)
	s.mb.RecordNatsServerCPUDataPoint(now, varz.CPU)
	s.mb.RecordNatsServerSlowConsumersDataPoint(now, varz.SlowConsumers)
	s.mb.RecordNatsServerRoutesDataPoint(now, varz.Routes)
	s.mb.RecordNatsServerLeafNodesDataPoint(now, varz.LeafNodes)
	s.mb.RecordNatsServerMaxConnectionsDataPoint(now, varz.MaxConnections)

	return s.mb.Emit(), nil
}

func (s *serverScraper) fetchVarz(ctx context.Context) (*VarzResponse, error) {
	reqURL := s.cfg.ClientConfig.Endpoint + "/varz"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch /varz: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d from /varz", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read /varz response body: %w", err)
	}

	var varz VarzResponse
	if err := json.Unmarshal(body, &varz); err != nil {
		return nil, fmt.Errorf("failed to unmarshal /varz response: %w", err)
	}

	return &varz, nil
}

func createServerScraper(_ context.Context, cfg Config, settings receiver.Settings) (scraper.Metrics, error) {
	s := &serverScraper{
		cfg:      cfg,
		settings: settings,
	}
	return scraper.NewMetrics(
		s.scrape,
		scraper.WithStart(s.start),
	)
}
