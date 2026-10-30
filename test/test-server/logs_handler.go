package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type unifiedLogsSearchRequest struct {
	CRN    string         `json:"crn"`
	Source string         `json:"source"`
	Search map[string]any `json:"search"`
}

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
    {"name": "id", "type": "string", "default_column": false, "description": "Unique identifier for this log record."},
    {"name": "timestamp", "type": "timestamp", "default_column": true, "description": "Event time, RFC 3339 UTC."},
    {"name": "level", "type": "enum", "default_column": true, "description": "Severity level: INFO, WARN, ERROR."},
    {"name": "message", "type": "string", "default_column": true, "description": "Human-readable description of the event."},
    {"name": "crn", "type": "string", "default_column": false, "description": "Confluent Resource Name of the resource the event concerns."},
    {"name": "source", "type": "string", "default_column": false, "description": "Component or subsystem that generated the event."},
    {"name": "event_type", "type": "enum", "default_column": true, "description": "Type of cluster health event."},
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

// Handler for: "/logs/v1/sources/{name}"
func handleLogsSource(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		name := mux.Vars(r)["name"]
		source, ok := logSources[name]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprintf(w, `{"errors":[{"status":"404","detail":"Unknown source: %s"}]}`, name)
			return
		}

		_, _ = w.Write([]byte(source))
	}
}

// clusterHealthEvents are the records returned for the cluster-health source, keyed by the cluster they belong to.
var clusterHealthEvents = []map[string]any{
	{"id": "evt-101", "timestamp": "2026-10-14T14:41:08Z", "level": "ERROR", "message": "Broker lkc-123 unreachable", "event_type": "SERVICE_ISSUE_DETECTED", "cluster_id": "lkc-123"},
	{"id": "evt-087", "timestamp": "2026-10-14T06:02:19Z", "level": "WARN", "message": "Scheduled maintenance window started", "event_type": "SERVICE_MAINTENANCE", "cluster_id": "lkc-456"},
	{"id": "evt-033", "timestamp": "2026-10-13T20:15:44Z", "level": "INFO", "message": "Cluster lkc-789 status updated", "event_type": "CONFLUENT_STATUS_UPDATED", "cluster_id": "lkc-789"},
}

// handleUnifiedLogsSearch serves POST /logs/v1/search requests that name a log source.
func handleUnifiedLogsSearch(t *testing.T, w http.ResponseWriter, req unifiedLogsSearchRequest) {
	var data []map[string]any

	switch req.Source {
	case "cluster-health":
		for _, event := range clusterHealthEvents {
			// A resource-scoped CRN ends in kafka=<cluster>; broader scopes match every cluster.
			if strings.Contains(req.CRN, "/kafka=") && !strings.HasSuffix(req.CRN, "/kafka="+event["cluster_id"].(string)) {
				continue
			}
			if !searchMatches(req.Search, event) {
				continue
			}
			data = append(data, event)
		}
	case "connector":
		if strings.HasSuffix(req.CRN, "/connector=az-connector") {
			data = []map[string]any{
				{"timestamp": "2026-10-14T09:12:03Z", "level": "ERROR", "task_id": "task-0", "message": "Connection to upstream host was lost.", "exception.stacktrace": "org.apache.kafka.connect.errors.ConnectException", "id": "lcc-123", "name": "az-connector"},
			}
		}
	}

	if data == nil {
		data = []map[string]any{}
	}

	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"data": data, "api_version": "v1", "kind": "Logs"}))
}

func searchMatches(search map[string]any, event map[string]any) bool {
	for field, want := range search {
		if field == "search_text" {
			if !strings.Contains(fmt.Sprint(event["message"]), fmt.Sprint(want)) {
				return false
			}
			continue
		}

		var wanted []string
		switch v := want.(type) {
		case string:
			wanted = []string{v}
		case []any:
			for _, item := range v {
				wanted = append(wanted, fmt.Sprint(item))
			}
		}
		if !slices.Contains(wanted, fmt.Sprint(event[field])) {
			return false
		}
	}
	return true
}
