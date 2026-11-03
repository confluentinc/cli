package testserver

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/gorilla/mux"
)

var logSources = map[string]string{
	"cluster-health": `{
  "name": "cluster-health",
  "display_name": "Cluster health events",
  "resource_type": "kafka",
  "id_prefix": "lkc",
  "crn_elements": ["organization", "environment", "cloud-cluster", "kafka"],
  "allowed_scopes": ["ORGANIZATION", "ENVIRONMENT", "RESOURCE"],
  "search_fields": [
    {"name": "id", "type": "string"},
    {"name": "event_type", "type": "enum", "allowed_values": ["SERVICE_MAINTENANCE", "HEALTH_ALERT", "CONFLUENT_STATUS_UPDATED", "SERVICE_ISSUE_DETECTED"]},
    {"name": "level", "type": "enum", "allowed_values": ["INFO", "WARN", "ERROR"]}
  ],
  "full_text_search_supported": true,
  "full_text_search_fields": ["message"],
  "returned_fields": [
    {"name": "id", "type": "string", "default_column": false, "description": "Unique identifier for this log record."},
    {"name": "timestamp", "type": "timestamp", "default_column": true, "description": "Event time, RFC 3339 UTC."},
    {"name": "level", "type": "enum", "default_column": true, "description": "Severity level: INFO, WARN, ERROR."},
    {"name": "message", "type": "string", "default_column": true, "description": "Human-readable description of the event."},
    {"name": "crn", "type": "string", "default_column": false, "description": "Confluent Resource Name of the resource the event concerns."},
    {"name": "source", "type": "string", "default_column": false, "description": "Component or subsystem that generated the event."},
    {"name": "event_type", "type": "enum", "default_column": true, "description": "Type of cluster health event."},
    {"name": "event_id", "type": "string", "default_column": false, "description": "Event-specific identifier, when provided."},
    {"name": "status", "type": "string", "default_column": false, "description": "Status associated with the event, when applicable."},
    {"name": "cloud_cluster_id", "type": "string", "default_column": false, "description": "lkc- ID the event belongs to."}
  ],
  "max_lookback_days": 3,
  "kind": "Source",
  "api_version": "logging/v1"
}`,
}

// logSourceRequests counts the discovery requests, for one source or the list, so tests can tell a cached source from a fetched one.
var logSourceRequests atomic.Int64

// LogSourceRequests returns how many log source discovery requests the test server has served.
func LogSourceRequests() int64 {
	return logSourceRequests.Load()
}

// LogSourcesOverride replaces the response of the log source discovery endpoints, so tests can reach the API's error responses and
// source lists the default mock does not serve.
type LogSourcesOverride struct {
	Status int
	Header map[string]string
	Body   string
}

var logSourcesOverride atomic.Pointer[LogSourcesOverride]

// SetLogSourcesOverride makes "/logs/v1/sources" and "/logs/v1/sources/{name}" answer with the given response. A nil override restores the default.
func SetLogSourcesOverride(override *LogSourcesOverride) {
	logSourcesOverride.Store(override)
}

// serveLogSourcesOverride answers with the override, if a test set one, and reports whether it did.
func serveLogSourcesOverride(w http.ResponseWriter) bool {
	override := logSourcesOverride.Load()
	if override == nil {
		return false
	}

	for key, value := range override.Header {
		w.Header().Set(key, value)
	}
	w.WriteHeader(override.Status)
	_, _ = fmt.Fprint(w, override.Body)
	return true
}

// Handler for: "/logs/v1/sources"
func handleLogsSources(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logSourceRequests.Add(1)

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if serveLogSourcesOverride(w) {
			return
		}

		names := make([]string, 0, len(logSources))
		for name := range logSources {
			names = append(names, name)
		}
		slices.Sort(names)

		data := make([]string, len(names))
		for i, name := range names {
			data[i] = logSources[name]
		}

		_, _ = fmt.Fprintf(w, `{"api_version":"logging/v1","kind":"SourceList","metadata":{"total_size":%d},"data":[%s]}`, len(data), strings.Join(data, ","))
	}
}

// Handler for: "/logs/v1/sources/{name}"
func handleLogsSource(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logSourceRequests.Add(1)

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		name := mux.Vars(r)["name"]
		if name == "rate-limited" {
			w.Header().Set("Retry-After", "12")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = fmt.Fprint(w, `{"errors":[{"status":"429","detail":"Too many source discovery requests. Please see the documentation for current rate limits."}]}`)
			return
		}

		source, ok := logSources[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, `{"errors":[{"status":"404","detail":"Unknown source: %s"}]}`, name)
			return
		}

		_, _ = w.Write([]byte(source))
	}
}
