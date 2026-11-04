package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
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

// LogSourceRequests counts the discovery requests, for one source or the list, so tests can tell a cached source from a fetched one.
var LogSourceRequests = 0

// LogSourcesResponse replaces the response of the log source discovery endpoints, so tests can reach the API's error responses and
// source lists the default mock does not serve.
type LogSourcesResponse struct {
	Status int
	Header map[string]string
	Body   string
}

// LogSourcesOverride, when set, is the response of "/logs/v1/sources" and "/logs/v1/sources/{name}". Tests reset it with defer.
var LogSourcesOverride *LogSourcesResponse

// LogsSearchOverride, when set, is the response of "/logs/v1/search" for requests that name a log source. Tests reset it with defer.
var LogsSearchOverride *LogSourcesResponse

// serveOverride answers with the override, if a test set one, and reports whether it did.
func serveOverride(w http.ResponseWriter, override *LogSourcesResponse) bool {
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
		LogSourceRequests++

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if serveOverride(w, LogSourcesOverride) {
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
		LogSourceRequests++

		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		name := mux.Vars(r)["name"]
		if serveOverride(w, LogSourcesOverride) {
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

type unifiedLogsSearchRequest struct {
	CRN       string         `json:"crn"`
	Source    string         `json:"source"`
	Search    map[string]any `json:"search"`
	Sort      string         `json:"sort"`
	StartTime string         `json:"start_time"`
	EndTime   string         `json:"end_time"`
}

// clusterHealthEvents are the records returned for the cluster-health source, keyed by the cluster they belong to.
var clusterHealthEvents = []map[string]any{
	{"id": "evt-101", "timestamp": "2026-10-14T14:41:08Z", "level": "ERROR", "message": "Broker lkc-123 unreachable", "event_type": "SERVICE_ISSUE_DETECTED", "cloud_cluster_id": "lkc-123"},
	{"id": "evt-087", "timestamp": "2026-10-14T06:02:19Z", "level": "WARN", "message": "Scheduled maintenance window started", "event_type": "SERVICE_MAINTENANCE", "cloud_cluster_id": "lkc-456"},
	{"id": "evt-033", "timestamp": "2026-10-13T20:15:44Z", "level": "INFO", "message": "Cluster lkc-789 status updated", "event_type": "CONFLUENT_STATUS_UPDATED", "cloud_cluster_id": "lkc-789"},
}

// pagedClusterEvents belong to the one cluster the mock serves in pages of two, to exercise `logs fetch --next`.
var pagedClusterEvents = []map[string]any{
	{"id": "evt-203", "timestamp": "2026-10-14T13:05:00Z", "level": "ERROR", "message": "Broker lkc-777 unreachable", "event_type": "SERVICE_ISSUE_DETECTED", "cloud_cluster_id": "lkc-777"},
	{"id": "evt-202", "timestamp": "2026-10-14T12:05:00Z", "level": "WARN", "message": "Broker lkc-777 slow", "event_type": "HEALTH_ALERT", "cloud_cluster_id": "lkc-777"},
	{"id": "evt-201", "timestamp": "2026-10-14T11:05:00Z", "level": "INFO", "message": "Cluster lkc-777 status updated", "event_type": "CONFLUENT_STATUS_UPDATED", "cloud_cluster_id": "lkc-777"},
}

const pagedClusterPageSize = 2

// handleUnifiedLogsSearch serves POST /logs/v1/search requests that name a log source.
func handleUnifiedLogsSearch(t *testing.T, w http.ResponseWriter, r *http.Request, req unifiedLogsSearchRequest) {
	w.Header().Set("Content-Type", "application/json")
	if serveOverride(w, LogsSearchOverride) {
		return
	}

	// The CLI must send a well-formed, non-empty time window, newest-first ordering and an explicit page size.
	start, err := time.Parse(time.RFC3339Nano, req.StartTime)
	require.NoError(t, err, "start_time")
	end, err := time.Parse(time.RFC3339Nano, req.EndTime)
	require.NoError(t, err, "end_time")
	require.False(t, end.Before(start), "end_time must not be before start_time")
	require.Equal(t, "desc", req.Sort)

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	require.NoError(t, err, "page_size")
	require.Positive(t, pageSize)

	// The scope of the query shows in which CRN elements are present.
	require.True(t, strings.HasPrefix(req.CRN, "crn://"), req.CRN)
	require.Contains(t, req.CRN, "/organization=", req.CRN)

	// page_token is the offset of the first record on the page.
	offset := 0
	if token := r.URL.Query().Get("page_token"); token != "" {
		offset, err = strconv.Atoi(token)
		require.NoError(t, err, "page_token")
	}

	var data []map[string]any

	switch req.Source {
	case "cluster-health":
		if strings.HasSuffix(req.CRN, "/kafka=lkc-777") {
			pageSize = min(pageSize, pagedClusterPageSize)
			data = pagedClusterEvents
			break
		}
		// An organization-scoped CRN has no environment element, and an environment-scoped one has no cluster.
		if strings.Contains(req.CRN, "/kafka=") {
			require.Contains(t, req.CRN, "/environment=", req.CRN)
		}
		for _, event := range clusterHealthEvents {
			// A resource-scoped CRN ends in kafka=<cluster>; broader scopes match every cluster.
			if strings.Contains(req.CRN, "/kafka=") && !strings.HasSuffix(req.CRN, "/kafka="+event["cloud_cluster_id"].(string)) {
				continue
			}
			if !searchMatches(req.Search, event) {
				continue
			}
			data = append(data, event)
		}
	}

	if data == nil {
		data = []map[string]any{}
	}
	response := map[string]any{"api_version": "v1", "kind": "Logs"}
	data = data[min(offset, len(data)):]
	if len(data) > pageSize {
		data = data[:pageSize]
		response["metadata"] = map[string]any{"next": fmt.Sprintf("http://%s%s?page_size=%d&page_token=%d", r.Host, r.URL.Path, pageSize, offset+pageSize)}
	}
	response["data"] = data

	w.Header().Set("Content-Type", "application/json")
	require.NoError(t, json.NewEncoder(w).Encode(response))
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
