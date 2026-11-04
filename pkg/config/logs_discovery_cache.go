package config

import "time"

// LogsDiscoveryCache keeps log source definitions from the logging API's discovery endpoints, so that repeated
// `confluent logs` commands and shell completion do not call discovery every time. The definitions are stored as the
// API returned them, which keeps this package free of the SDK's types.
type LogsDiscoveryCache struct {
	// Key is the platform server and organization the entries were fetched for.
	Key string `json:"key,omitempty"`
	// ListedAt is when the full list of log sources was last fetched. It is zero when only single sources are cached,
	// which is not enough to answer "which log sources exist".
	ListedAt time.Time `json:"listed_at,omitempty"`
	// Sources holds each log source by name.
	Sources map[string]*LogsCachedSource `json:"sources,omitempty"`
}

type LogsCachedSource struct {
	FetchedAt time.Time      `json:"fetched_at"`
	Source    map[string]any `json:"source"`
}
