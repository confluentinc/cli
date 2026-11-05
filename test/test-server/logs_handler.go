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

// LastLogsSearchCRN is the CRN of the latest search request that named a log source, so tests can check which scope the
// CLI asked for.
var LastLogsSearchCRN string

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

// clusterHealthEvent builds a record in the shape the cluster-health source returns: the cluster ID is both the
// record's id and its cloud_cluster_id, and the CRN names the cluster.
func clusterHealthEvent(clusterId, eventId, level, eventType, message, status, timestamp string) map[string]any {
	return map[string]any{
		"cloud_cluster_id": clusterId,
		"event_type":       eventType,
		"event_id":         eventId,
		"level":            level,
		"source":           "ClusterHealth",
		"id":               clusterId,
		"message":          message,
		"crn":              "crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=" + clusterId + "/kafka=" + clusterId,
		"status":           status,
		"timestamp":        timestamp,
	}
}

// clusterHealthEvents are the records returned for the cluster-health source, keyed by the cluster they belong to.
var clusterHealthEvents = []map[string]any{
	clusterHealthEvent("lkc-123", "4060c2e9-b415-4b0f-a11f-359177689d8c", "ERROR", "SERVICE_ISSUE_DETECTED", "Broker lkc-123 unreachable", "OPEN", "2026-10-14T14:41:08.025Z"),
	clusterHealthEvent("lkc-456", "7c1f6a52-0d3e-4c8b-9a41-2b5e8d9f0a13", "WARN", "SERVICE_MAINTENANCE", "Scheduled maintenance window started", "OPEN", "2026-10-14T06:02:19.310Z"),
	clusterHealthEvent("lkc-789", "e2b94d17-58aa-4f0c-b6d3-91c7a0e4f825", "INFO", "CONFLUENT_STATUS_UPDATED", "Cluster lkc-789 status updated", "COMPLETED", "2026-10-13T20:15:44.871Z"),
}

// pagedClusterEvents belong to the one cluster the mock serves in pages of two, to exercise `logs fetch --next`.
var pagedClusterEvents = []map[string]any{
	clusterHealthEvent("lkc-777", "1d0a3f64-92b7-4e15-8c2a-6f4b7e9d1a30", "ERROR", "SERVICE_ISSUE_DETECTED", "Broker lkc-777 unreachable", "OPEN", "2026-10-14T13:05:00.120Z"),
	clusterHealthEvent("lkc-777", "5a8e2c91-3f60-47d2-b1c4-0e9a6d7f3b58", "WARN", "HEALTH_ALERT", "Broker lkc-777 slow", "OPEN", "2026-10-14T12:05:00.480Z"),
	clusterHealthEvent("lkc-777", "9f3b7d20-61c4-4a8e-95d0-3c2e1b8a6f47", "INFO", "CONFLUENT_STATUS_UPDATED", "Cluster lkc-777 status updated", "COMPLETED", "2026-10-14T11:05:00.905Z"),
}

// connectorLogs are the records returned for the connector source, in the fields its deployed catalog entry returns. A
// record has no exception.stacktrace when the log line carries none.
var connectorLogs = []map[string]any{
	{"timestamp": "2026-10-14T09:12:03Z", "level": "ERROR", "task_id": "task-0", "message": "Connection to upstream host was lost.", "exception.stacktrace": "org.apache.kafka.connect.errors.ConnectException: Connection to upstream host was lost.", "id": "lcc-123", "name": "my-connector"},
	{"timestamp": "2026-10-14T08:50:41Z", "level": "WARN", "task_id": "task-1", "message": "Retrying connection, attempt 3 of 5.", "id": "lcc-123", "name": "my-connector"},
}

const pagedClusterPageSize = 2

// handleUnifiedLogsSearch serves POST /logs/v1/search requests that name a log source.
func handleUnifiedLogsSearch(t *testing.T, w http.ResponseWriter, r *http.Request, req unifiedLogsSearchRequest) {
	LastLogsSearchCRN = req.CRN

	w.Header().Set("Content-Type", "application/json")
	if serveOverride(w, LogsSearchOverride) {
		return
	}

	// The CLI must send a well-formed, non-empty time window, newest-first ordering and an explicit page size.
	start, err := time.Parse(time.RFC3339Nano, req.StartTime)
	require.NoError(t, err, "start_time")
	// The API documents RFC 3339 UTC with a Z suffix, and the CLI sends whole seconds.
	require.True(t, strings.HasSuffix(req.StartTime, "Z") && strings.HasSuffix(req.EndTime, "Z"), "timestamps must be UTC")
	require.NotContains(t, req.StartTime+req.EndTime, ".", "timestamps must be whole seconds")
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
	case "connector":
		for _, record := range connectorLogs {
			if strings.HasSuffix(req.CRN, "/connector="+record["id"].(string)) && searchMatches(req.Search, record) {
				data = append(data, record)
			}
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
