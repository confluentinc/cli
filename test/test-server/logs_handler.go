package testserver

import (
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
)

var logSources = map[string]string{
	"cluster-health": `{
  "api_version": "logging/v1",
  "kind": "Source",
  "name": "cluster-health",
  "display_name": "Cluster health events",
  "resource_type": "kafka",
  "id_prefix": "lkc",
  "crn_elements": ["organization", "environment", "cloud-cluster", "kafka"],
  "allowed_scopes": ["ORGANIZATION", "ENVIRONMENT", "RESOURCE"],
  "search_fields": [
    {"name": "id", "type": "string"},
    {"name": "event_type", "type": "enum", "allowed_values": ["SERVICE_MAINTENANCE", "HEALTH_ALERT", "CONFLUENT_STATUS_UPDATED", "SERVICE_ISSUE_DETECTED"]},
    {"name": "level", "type": "enum", "allowed_values": ["INFO", "WARN", "ERROR"]},
    {"name": "client_id", "type": "string"}
  ],
  "full_text_search_supported": true,
  "full_text_search_fields": ["message"],
  "returned_fields": [
    {"name": "timestamp", "type": "timestamp", "default_column": true, "description": "Event time, RFC 3339 UTC."},
    {"name": "level", "type": "enum", "default_column": true, "description": "Severity level: INFO, WARN, ERROR."},
    {"name": "event_type", "type": "enum", "default_column": true, "description": "Type of cluster health event."},
    {"name": "message", "type": "string", "default_column": true, "description": "Human-readable description of the event."},
    {"name": "id", "type": "string", "default_column": false, "description": "Unique identifier for this log record."},
    {"name": "crn", "type": "string", "default_column": false, "description": "Confluent Resource Name of the resource the event concerns."},
    {"name": "source", "type": "string", "default_column": false, "description": "Component or subsystem that generated the event."},
    {"name": "event_id", "type": "string", "default_column": false, "description": "Event-specific identifier, when provided."},
    {"name": "status", "type": "string", "default_column": false, "description": "Status associated with the event, when applicable."},
    {"name": "client_id", "type": "string", "default_column": false, "description": "Kafka client ID, when the event names one."},
    {"name": "cluster_id", "type": "string", "default_column": false, "description": "lkc- ID the event belongs to."}
  ],
  "max_lookback_days": 3
}`,
	"connector": `{
  "api_version": "logging/v1",
  "kind": "Source",
  "name": "connector",
  "display_name": "Connector logs",
  "resource_type": "connector",
  "crn_elements": ["organization", "environment", "cloud-cluster", "connector"],
  "allowed_scopes": ["RESOURCE"],
  "search_fields": [
    {"name": "level", "type": "enum", "allowed_values": ["INFO", "WARN", "ERROR", "FATAL"]},
    {"name": "task_id", "type": "string"},
    {"name": "id", "type": "string"}
  ],
  "full_text_search_supported": true,
  "full_text_search_fields": ["message", "exception.stacktrace"],
  "returned_fields": [
    {"name": "timestamp", "type": "timestamp", "default_column": true},
    {"name": "level", "type": "enum", "default_column": true},
    {"name": "task_id", "type": "string", "default_column": true},
    {"name": "message", "type": "string", "default_column": true},
    {"name": "exception.stacktrace", "type": "string", "default_column": true},
    {"name": "id", "type": "string", "default_column": true},
    {"name": "name", "type": "string", "default_column": true}
  ],
  "max_lookback_days": 3
}`,
}

// Handler for: "/logs/v1/sources"
func handleLogsSources(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

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
