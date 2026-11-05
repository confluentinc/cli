package logs

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// suggestionsOf returns the suggestions attached to an error, or "" if it carries none.
func suggestionsOf(err error) string {
	if withSuggestions, ok := err.(errors.ErrorWithSuggestions); ok {
		return withSuggestions.GetSuggestionsMsg()
	}
	return ""
}

var clusterHealth = &loggingv1.LoggingV1Source{
	Name:                    "cluster-health",
	ResourceType:            "kafka",
	IdPrefix:                loggingv1.PtrString("lkc"),
	CrnElements:             []string{"organization", "environment", "cloud-cluster", "kafka"},
	AllowedScopes:           []string{"ORGANIZATION", "ENVIRONMENT", "RESOURCE"},
	FullTextSearchSupported: true,
	MaxLookbackDays:         3,
	SearchFields: []loggingv1.LoggingV1SearchFieldDescriptor{
		{Name: "id", Type: "string"},
		{Name: "level", Type: "enum", AllowedValues: &[]string{"INFO", "WARN", "ERROR"}},
	},
}

// resourceScopedSource is a log source that is only queried per resource and declares no ID prefix.
var resourceScopedSource = &loggingv1.LoggingV1Source{
	Name:          "resource-source",
	ResourceType:  "example",
	CrnElements:   []string{"organization", "environment", "cloud-cluster", "example"},
	AllowedScopes: []string{"RESOURCE"},
}

func TestResolveTimeWindow(t *testing.T) {
	now := time.Date(2026, 10, 14, 15, 0, 0, 0, time.UTC)

	start, end, err := resolveTimeWindow("24h", "", "", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-13T15:00:00Z", start)
	require.Equal(t, "2026-10-14T15:00:00Z", end)

	start, _, err = resolveTimeWindow("7d", "", "", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-07T15:00:00Z", start)

	start, end, err = resolveTimeWindow("", "2026-10-14T10:00:00Z", "", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-14T10:00:00Z", start)
	require.Equal(t, "2026-10-14T15:00:00Z", end)

	_, _, err = resolveTimeWindow("1h", "2026-10-14T10:00:00Z", "", now)
	require.EqualError(t, err, "`--since` cannot be used with `--from` or `--to`")

	_, _, err = resolveTimeWindow("", "", "", now)
	require.EqualError(t, err, "one of `--since` or `--from` is required")

	_, _, err = resolveTimeWindow("", "2026-10-14T10:00:00", "", now)
	require.ErrorContains(t, err, "invalid `--from` value")
	require.Contains(t, suggestionsOf(err), "ISO 8601 UTC timestamp")

	_, _, err = resolveTimeWindow("", "2026-10-14T10:00:00Z", "2026-10-14T09:00:00Z", now)
	require.EqualError(t, err, "`--to` must not be before `--from`")

	// The API takes whole seconds, so a start is rounded down and an explicit end is rounded up. The window never shrinks.
	start, end, err = resolveTimeWindow("", "2026-10-14T10:00:00.250Z", "2026-10-14T10:00:00.750Z", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-14T10:00:00Z", start)
	require.Equal(t, "2026-10-14T10:00:01Z", end)

	start, end, err = resolveTimeWindow("500ms", "", "", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-14T14:59:59Z", start)
	require.Equal(t, "2026-10-14T15:00:00Z", end)

	_, end, err = resolveTimeWindow("", "2026-10-14T10:00:00Z", "2026-10-14T10:00:05Z", now)
	require.NoError(t, err)
	require.Equal(t, "2026-10-14T10:00:05Z", end, "a whole-second end is unchanged")

	// A day count that overflows time.Duration is rejected instead of wrapping around.
	for _, bad := range []string{"abc", "0h", "-1h", "0d", "xd", "213504d", "106752d", "9223372036854775807d"} {
		_, _, err = resolveTimeWindow(bad, "", "", now)
		require.ErrorContains(t, err, "invalid `--since` value", bad)
	}
}

func TestBuildSearch(t *testing.T) {
	search, err := buildSearch(nil, "")
	require.NoError(t, err)
	require.Nil(t, search)

	search, err = buildSearch([]string{"level=ERROR", "event_type=HEALTH_ALERT"}, "disk full")
	require.NoError(t, err)
	require.Equal(t, map[string]any{"level": "ERROR", "event_type": "HEALTH_ALERT", "search_text": "disk full"}, search)

	search, err = buildSearch([]string{"level=ERROR", "level=WARN"}, "")
	require.NoError(t, err)
	require.Equal(t, map[string]any{"level": []string{"ERROR", "WARN"}}, search)

	for _, bad := range []string{"level", "=ERROR", "level="} {
		_, err = buildSearch([]string{bad}, "")
		require.ErrorContains(t, err, "invalid `--search` value", bad)
	}

	_, err = buildSearch([]string{"search_text=x"}, "")
	require.ErrorContains(t, err, "reserved field")
	require.Contains(t, suggestionsOf(err), "`--search-text`")
}

func TestValidateSearch(t *testing.T) {
	require.NoError(t, validateSearch(clusterHealth, map[string]any{"level": "ERROR", "search_text": "x"}))
	require.NoError(t, validateSearch(clusterHealth, map[string]any{"level": []string{"ERROR", "WARN"}, "id": "evt-1"}))

	err := validateSearch(clusterHealth, map[string]any{"severity": "ERROR"})
	require.EqualError(t, err, `"severity" is not a search field for "cluster-health" logs`)
	require.Equal(t, "Available search fields: id, level.\nRun `confluent logs source describe cluster-health` for details.", suggestionsOf(err))

	err = validateSearch(clusterHealth, map[string]any{"level": "DEBUG"})
	require.EqualError(t, err, `"DEBUG" is not a valid value for "level"`)
	require.Equal(t, "Allowed values: INFO, WARN, ERROR.", suggestionsOf(err))

	noText := *clusterHealth
	noText.FullTextSearchSupported = false
	require.EqualError(t, validateSearch(&noText, map[string]any{"search_text": "x"}), "\"cluster-health\" logs do not support `--search-text`")
}

func TestScope(t *testing.T) {
	require.Equal(t, "RESOURCE", scopeFor("lkc-1", false))
	require.Equal(t, "RESOURCE", scopeFor("lkc-1", true))
	require.Equal(t, "ENVIRONMENT", scopeFor("", true))
	require.Equal(t, "ORGANIZATION", scopeFor("", false))

	require.NoError(t, validateScope(clusterHealth, "ORGANIZATION", ""))
	err := validateScope(resourceScopedSource, "ORGANIZATION", "")
	require.EqualError(t, err, `log source "resource-source" requires a resource ID`)
	require.Contains(t, suggestionsOf(err), "confluent logs fetch <resource-id> --source resource-source")
	err = validateScope(&loggingv1.LoggingV1Source{Name: "x", AllowedScopes: []string{"ENVIRONMENT"}}, "ORGANIZATION", "")
	require.EqualError(t, err, `log source "x" does not support organization scope`)
	require.Equal(t, "Supported scopes: environment.", suggestionsOf(err))
}

func TestValidateResourceId(t *testing.T) {
	require.NoError(t, validateResourceId(clusterHealth, "lkc-123"))
	require.NoError(t, validateResourceId(clusterHealth, ""))
	require.NoError(t, validateResourceId(resourceScopedSource, "anything")) // no id_prefix declared
	err := validateResourceId(clusterHealth, "lcc-123")
	require.EqualError(t, err, `resource ID "lcc-123" does not match log source "cluster-health"`)
	require.Equal(t, `This log source expects IDs starting with "lkc-".`, suggestionsOf(err))
}

func TestValidateLookback(t *testing.T) {
	require.NoError(t, validateLookback(clusterHealth, "72h"))
	require.NoError(t, validateLookback(clusterHealth, "3d"))
	require.NoError(t, validateLookback(clusterHealth, ""))
	err := validateLookback(clusterHealth, "30d")
	require.EqualError(t, err, "`--since` exceeds the 3d lookback for \"cluster-health\"")
	require.Equal(t, "Use a window of at most 3d.", suggestionsOf(err))
}

func TestBuildCRN(t *testing.T) {
	values := crnValues{Organization: "org-1", Environment: "env-1", Cluster: "lkc-1", Leaf: "lkc-1"}

	crn, err := buildCRN(clusterHealth, "RESOURCE", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1/environment=env-1/cloud-cluster=lkc-1/kafka=lkc-1", crn)

	crn, err = buildCRN(clusterHealth, "ENVIRONMENT", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1/environment=env-1", crn)

	crn, err = buildCRN(clusterHealth, "ORGANIZATION", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1", crn)

	values.Cluster, values.Leaf = "lkc-9", "my-resource"
	crn, err = buildCRN(resourceScopedSource, "RESOURCE", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1/environment=env-1/cloud-cluster=lkc-9/example=my-resource", crn)

	_, err = buildCRN(&loggingv1.LoggingV1Source{Name: "x", CrnElements: []string{"organization", "flink-region", "thing"}}, "RESOURCE", values)
	require.ErrorContains(t, err, `uses a CRN element "flink-region" that this version of the CLI does not support`)
}

// TestBuildCRNCloudClusterLeaf covers a source whose CRN ends at the cloud cluster. The last element always takes the
// resource ID, even when a separate cluster value is set, and the broader scopes stop earlier.
func TestBuildCRNCloudClusterLeaf(t *testing.T) {
	source := &loggingv1.LoggingV1Source{
		Name:        "cluster-source",
		CrnElements: []string{"organization", "environment", "cloud-cluster"},
	}
	values := crnValues{Organization: "org-1", Environment: "env-1", Cluster: "lkc-other", Leaf: "lkc-1"}

	crn, err := buildCRN(source, "RESOURCE", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1/environment=env-1/cloud-cluster=lkc-1", crn)

	crn, err = buildCRN(source, "ENVIRONMENT", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1/environment=env-1", crn)

	crn, err = buildCRN(source, "ORGANIZATION", values)
	require.NoError(t, err)
	require.Equal(t, "crn://confluent.cloud/organization=org-1", crn)

	// A source with no element beyond the scope cannot be scoped to it.
	short := &loggingv1.LoggingV1Source{Name: "org-only", CrnElements: []string{"organization"}}
	_, err = buildCRN(short, "ORGANIZATION", values)
	require.ErrorContains(t, err, `does not declare an organization element to scope a query to`)
}

func TestColumnHeader(t *testing.T) {
	require.Equal(t, "Timestamp", columnHeader("timestamp"))
	require.Equal(t, "Event Type", columnHeader("event_type"))
	require.Equal(t, "Task ID", columnHeader("task_id"))
	require.Equal(t, "ID", columnHeader("id"))
	require.Equal(t, "Exception Stacktrace", columnHeader("exception.stacktrace"))
}

func TestRecordValue(t *testing.T) {
	record := map[string]any{
		"level":                "ERROR",
		"exception.stacktrace": "flat",
		"nested":               map[string]any{"value": "deep"},
		"count":                float64(3),
	}
	require.Equal(t, "ERROR", recordValue(record, "level"))
	require.Equal(t, "flat", recordValue(record, "exception.stacktrace"))
	require.Equal(t, "deep", recordValue(record, "nested.value"))
	require.Equal(t, "3", recordValue(record, "count"))
	require.Equal(t, "", recordValue(record, "missing"))
}

func TestWriteLogsToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "logs.jsonl")

	first := []map[string]any{{"level": "ERROR", "message": "one"}, {"level": "WARN", "message": "two"}}
	require.NoError(t, writeLogsToFile(path, first))

	// A second call appends instead of overwriting.
	require.NoError(t, writeLogsToFile(path, []map[string]any{{"level": "INFO", "message": "three"}}))

	got, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, `{"level":"ERROR","message":"one"}
{"level":"WARN","message":"two"}
{"level":"INFO","message":"three"}
`, string(got))

	require.ErrorContains(t, writeLogsToFile(filepath.Join(t.TempDir(), "missing", "logs.jsonl"), first), "failed to open file")
}

func TestQueryKey(t *testing.T) {
	search := map[string]any{"level": []string{"ERROR", "WARN"}, "search_text": "boom"}

	key, err := queryKey("crn://a", "cluster-health", search, "24h", "", "")
	require.NoError(t, err)

	same, err := queryKey("crn://a", "cluster-health", map[string]any{"search_text": "boom", "level": []string{"ERROR", "WARN"}}, "24h", "", "")
	require.NoError(t, err)
	require.Equal(t, key, same)

	for name, other := range map[string][]any{
		"resource": {"crn://b", "cluster-health", search, "24h", "", ""},
		"source":   {"crn://a", "resource-source", search, "24h", "", ""},
		"search":   {"crn://a", "cluster-health", map[string]any{"level": "ERROR"}, "24h", "", ""},
		"since":    {"crn://a", "cluster-health", search, "1h", "", ""},
		"from":     {"crn://a", "cluster-health", search, "", "2026-10-14T00:00:00Z", ""},
	} {
		got, err := queryKey(other[0].(string), other[1].(string), other[2].(map[string]any), other[3].(string), other[4].(string), other[5].(string))
		require.NoError(t, err)
		require.NotEqual(t, key, got, name)
	}
}
