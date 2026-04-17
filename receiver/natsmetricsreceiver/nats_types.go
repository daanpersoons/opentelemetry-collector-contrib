// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package natsmetricsreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/natsmetricsreceiver"

// VarzResponse represents the response from the NATS /varz monitoring endpoint.
type VarzResponse struct {
	ServerID         string  `json:"server_id"`
	Version          string  `json:"version"`
	Connections      int64   `json:"connections"`
	TotalConnections int64   `json:"total_connections"`
	Routes           int64   `json:"routes"`
	LeafNodes        int64   `json:"leafnodes"`
	MaxConnections   int64   `json:"max_connections"`
	Subscriptions    int64   `json:"subscriptions"`
	InMsgs           int64   `json:"in_msgs"`
	OutMsgs          int64   `json:"out_msgs"`
	InBytes          int64   `json:"in_bytes"`
	OutBytes         int64   `json:"out_bytes"`
	SlowConsumers    int64   `json:"slow_consumers"`
	Mem              int64   `json:"mem"`
	CPU              float64 `json:"cpu"`
}

// ConnzResponse represents the response from the NATS /connz monitoring endpoint.
type ConnzResponse struct {
	NumConnections int64 `json:"num_connections"`
	Total          int64 `json:"total"`
}

// JszResponse represents the response from the NATS /jsz monitoring endpoint.
type JszResponse struct {
	Streams        int64  `json:"streams"`
	Consumers      int64  `json:"consumers"`
	Messages       int64  `json:"messages"`
	Bytes          int64  `json:"bytes"`
	ReservedMemory int64  `json:"reserved_memory"`
	ReservedStore  int64  `json:"reserved_store"`
	API            JszAPI `json:"api"`
}

// JszAPI represents JetStream API statistics.
type JszAPI struct {
	Total  int64 `json:"total"`
	Errors int64 `json:"errors"`
}

// AccStatzResponse represents the response from the NATS /accstatz monitoring endpoint.
type AccStatzResponse struct {
	AccountStats []AccountStat `json:"account_statz"`
}

// AccountStat represents statistics for a single NATS account.
type AccountStat struct {
	Acc           string    `json:"acc"`
	Conns         int64     `json:"conns"`
	LeafNodes     int64     `json:"leafnodes"`
	Subs          int64     `json:"subs"`
	Sent          DataStats `json:"sent"`
	Received      DataStats `json:"received"`
	SlowConsumers int64     `json:"slow_consumers"`
}

// DataStats represents byte/message statistics.
type DataStats struct {
	Msgs  int64 `json:"msgs"`
	Bytes int64 `json:"bytes"`
}
