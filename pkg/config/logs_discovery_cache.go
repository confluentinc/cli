package config

import (
	"encoding/json"
	"time"
)

// LogsDiscoveryCache keeps log source definitions from the logging API's discovery endpoints, so that repeated
// `confluent logs` commands and shell completion do not call discovery every time. The definitions are stored as the
// API returned them, which keeps this package free of the SDK's types.
type LogsDiscoveryCache struct {
	// Key is the platform server and organization the entries were fetched for.
	Key string `json:"key,omitempty"`
	// Sources holds each log source by name.
	Sources map[string]*LogsCachedSource `json:"sources,omitempty"`
}

type LogsCachedSource struct {
	FetchedAt time.Time       `json:"fetched_at"`
	Source    json.RawMessage `json:"source"`
}
