package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

func (s *CLITestSuite) TestLogsSourceList() {
	tests := []CLITest{
		{args: "logs source list", fixture: "logs/source/list.golden"},
		{args: "logs source list -o json", fixture: "logs/source/list-json.golden"},
		{args: "logs source list -o yaml", fixture: "logs/source/list-yaml.golden"},
		{args: "logs source list extra", fixture: "logs/source/list-extra-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestLogsSourceDescribe() {
	tests := []CLITest{
		{args: "logs source describe cluster-health", fixture: "logs/source/describe.golden"},
		{args: "logs source describe cluster-health -o json", fixture: "logs/source/describe-json.golden"},
		{args: "logs source describe cluster-health -o yaml", fixture: "logs/source/describe-yaml.golden"},
		{args: "logs source describe connector", fixture: "logs/source/describe-unknown.golden", exitCode: 1},
		{args: "logs source describe", fixture: "logs/source/describe-missing-arg.golden", exitCode: 1},
		{args: "logs source describe cluster-health connector", fixture: "logs/source/describe-extra-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestLogsAutocomplete() {
	tests := []CLITest{
		{args: `__complete logs source describe ""`, fixture: "logs/source/describe-autocomplete.golden"},
		{args: `__complete logs source describe cluster-health ""`, fixture: "logs/source/describe-autocomplete-second-arg.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

// TestLogsSourceListRefreshesDiscovery checks that `source list` always asks the API.
func (s *CLITestSuite) TestLogsSourceListRefreshesDiscovery() {
	before := testserver.LogSourceRequests
	discoveryRequests := func(t *testing.T, want int) {
		require.Equal(t, want, testserver.LogSourceRequests-before)
	}

	s.runIntegrationTest(CLITest{login: "cloud", args: "logs source list", fixture: "logs/source/list.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: "logs source list", fixture: "logs/source/list.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 2) }})
}

// TestLogsSourceDescribeRefreshesDiscovery checks that `source describe` always asks the API, and that completion
// reuses the list `source list` fetched, even after `source describe` cached a single source.
func (s *CLITestSuite) TestLogsSourceDescribeRefreshesDiscovery() {
	before := testserver.LogSourceRequests
	discoveryRequests := func(t *testing.T, want int) {
		require.Equal(t, want, testserver.LogSourceRequests-before)
	}

	s.runIntegrationTest(CLITest{login: "cloud", args: "logs source list", fixture: "logs/source/list.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: `__complete logs source describe ""`, fixture: "logs/source/describe-autocomplete.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: "logs source describe cluster-health", fixture: "logs/source/describe.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 2) }})
	s.runIntegrationTest(CLITest{workflow: true, args: `__complete logs source describe ""`, fixture: "logs/source/describe-autocomplete.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 2) }})
	s.runIntegrationTest(CLITest{workflow: true, args: "logs source describe cluster-health", fixture: "logs/source/describe.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 3) }})
}

func (s *CLITestSuite) TestLogsSourceListNotLoggedIn() {
	s.runIntegrationTest(CLITest{args: "logs source list", fixture: "logs/source/list-not-logged-in.golden", exitCode: 1})
}

// logSourcesResponses are the responses of GET /logs/v1/sources that the default mock does not serve: sources of other
// shapes, and each error the LLD lists for the endpoint.
var logSourcesResponses = map[string]*testserver.LogSourcesResponse{
	"two-sources": {Status: 200, Body: `{"api_version":"logging/v1","kind":"SourceList","metadata":{"total_size":2},"data":[
  {"api_version":"logging/v1","kind":"Source","name":"cluster-health","display_name":"Cluster health events","resource_type":"kafka","id_prefix":"lkc","crn_elements":["organization","environment","cloud-cluster","kafka"],"allowed_scopes":["ORGANIZATION","ENVIRONMENT","RESOURCE"],"search_fields":[],"full_text_search_supported":true,"returned_fields":[],"max_lookback_days":3},
  {"api_version":"logging/v1","kind":"Source","name":"cluster-link-insights","display_name":"Cluster Linking insights","resource_type":"cluster-link","crn_elements":["organization","environment","cloud-cluster","cluster-link"],"allowed_scopes":["RESOURCE"],"search_fields":[],"full_text_search_supported":false,"returned_fields":[],"max_lookback_days":7}]}`},
	"empty":                  {Status: 200, Body: `{"api_version":"logging/v1","kind":"SourceList","metadata":{"total_size":0},"data":[]}`},
	"unauthenticated":        {Status: 401, Body: `{"errors":[{"status":"401","detail":"Credentials are required to access this resource."}]}`},
	"feature-not-enabled":    {Status: 403, Body: `{"errors":[{"status":"403","detail":"Feature is not enabled for this organization"}]}`},
	"organization-required":  {Status: 403, Body: `{"errors":[{"status":"403","detail":"Organization ID is required"}]}`},
	"rate-limited":           {Status: 429, Header: map[string]string{"Retry-After": "12"}, Body: `{"errors":[{"status":"429","detail":"Too many source discovery requests. Please see the documentation for current rate limits."}]}`},
	"rate-limited-no-header": {Status: 429, Body: `{"errors":[{"status":"429","detail":"Too many source discovery requests. Please see the documentation for current rate limits."}]}`},
	"rate-limited-per-ip":    {Status: 429, Header: map[string]string{"Retry-After": "1"}, Body: `{"errors":[{"status":"429","detail":"Too many requests have been made from this IP address: 203.0.113.7. Please see the documentation for current rate limits."}]}`},
	"server-error":           {Status: 500, Body: `{"errors":[{"status":"500","detail":"Internal server error"}]}`},
}

// TestLogsSourceListResponses covers the other responses of GET /logs/v1/sources: sources of other shapes, and every
// error the LLD lists for the endpoint.
func (s *CLITestSuite) TestLogsSourceListResponses() {
	tests := []struct {
		response string
		test     CLITest
	}{
		{"two-sources", CLITest{args: "logs source list", fixture: "logs/source/list-two-sources.golden"}},
		{"two-sources", CLITest{args: "logs source list -o json", fixture: "logs/source/list-two-sources-json.golden"}},
		{"empty", CLITest{args: "logs source list", fixture: "logs/source/list-empty.golden"}},
		{"empty", CLITest{args: "logs source list -o json", fixture: "logs/source/list-empty-json.golden"}},
		{"unauthenticated", CLITest{args: "logs source list", fixture: "logs/source/list-unauthenticated.golden", exitCode: 1}},
		{"feature-not-enabled", CLITest{args: "logs source list", fixture: "logs/source/list-feature-not-enabled.golden", exitCode: 1}},
		{"organization-required", CLITest{args: "logs source list", fixture: "logs/source/list-organization-required.golden", exitCode: 1}},
		{"rate-limited", CLITest{args: "logs source list", fixture: "logs/source/list-rate-limited.golden", exitCode: 1}},
		{"rate-limited-no-header", CLITest{args: "logs source list", fixture: "logs/source/list-rate-limited-no-header.golden", exitCode: 1}},
		{"rate-limited-per-ip", CLITest{args: "logs source list", fixture: "logs/source/list-rate-limited-per-ip.golden", exitCode: 1}},
		{"server-error", CLITest{args: "logs source list", fixture: "logs/source/list-server-error.golden", exitCode: 1}},
	}

	defer func() { testserver.LogSourcesOverride = nil }()
	for _, tt := range tests {
		testserver.LogSourcesOverride = logSourcesResponses[tt.response]
		tt.test.login = "cloud"
		s.runIntegrationTest(tt.test)
	}
}

func (s *CLITestSuite) TestLogsSourceDescribeNotLoggedIn() {
	s.runIntegrationTest(CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-not-logged-in.golden", exitCode: 1})
}

// TestLogsSourceDescribeResponses covers the other responses of GET /logs/v1/sources/{name}: a source of another
// shape, and every error the LLD lists for the endpoint.
func (s *CLITestSuite) TestLogsSourceDescribeResponses() {
	responses := map[string]*testserver.LogSourcesResponse{
		"plain-source": {Status: 200, Body: `{"api_version":"logging/v1","kind":"Source","name":"cluster-link-insights","display_name":"Cluster Linking insights","resource_type":"cluster-link","crn_elements":["organization","environment","cloud-cluster","cluster-link"],"allowed_scopes":["RESOURCE"],"search_fields":[],"full_text_search_supported":false,"returned_fields":[{"name":"timestamp","type":"timestamp","default_column":true},{"name":"message","type":"string","default_column":true}],"max_lookback_days":7}`},
	}
	for _, name := range []string{"unauthenticated", "feature-not-enabled", "organization-required", "rate-limited", "rate-limited-no-header", "rate-limited-per-ip", "server-error"} {
		responses[name] = logSourcesResponses[name]
	}

	tests := []struct {
		response string
		test     CLITest
	}{
		{"plain-source", CLITest{args: "logs source describe cluster-link-insights", fixture: "logs/source/describe-plain-source.golden"}},
		{"plain-source", CLITest{args: "logs source describe cluster-link-insights -o json", fixture: "logs/source/describe-plain-source-json.golden"}},
		{"unauthenticated", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-unauthenticated.golden", exitCode: 1}},
		{"feature-not-enabled", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-feature-not-enabled.golden", exitCode: 1}},
		{"organization-required", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-organization-required.golden", exitCode: 1}},
		{"rate-limited", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-rate-limited.golden", exitCode: 1}},
		{"rate-limited-no-header", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-rate-limited-no-header.golden", exitCode: 1}},
		{"rate-limited-per-ip", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-rate-limited-per-ip.golden", exitCode: 1}},
		{"server-error", CLITest{args: "logs source describe cluster-health", fixture: "logs/source/describe-server-error.golden", exitCode: 1}},
	}

	defer func() { testserver.LogSourcesOverride = nil }()
	for _, tt := range tests {
		testserver.LogSourcesOverride = responses[tt.response]
		tt.test.login = "cloud"
		s.runIntegrationTest(tt.test)
	}
}

func (s *CLITestSuite) TestLogsFetch() {
	tests := []CLITest{
		// Resource scope
		{args: "logs fetch lkc-123 --source cluster-health --since 1h", fixture: "logs/fetch/resource.golden"},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h -o json", fixture: "logs/fetch/resource-json.golden"},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h -o yaml", fixture: "logs/fetch/resource-yaml.golden"},
		{args: "logs fetch lkc-456 --source cluster-health --from 2026-10-14T00:00:00Z --to 2026-10-14T23:59:59Z --environment env-596", fixture: "logs/fetch/resource-from-to.golden"},
		{args: "logs fetch lkc-123 --source cluster-health --since 3d --search level=ERROR --search event_type=SERVICE_ISSUE_DETECTED --tail 100", fixture: "logs/fetch/resource-search.golden"},
		{args: "logs fetch lkc-456 --source cluster-health --since 3d --search level=WARN,level=ERROR --search-text maintenance", fixture: "logs/fetch/resource-search-text.golden"},
		{args: "logs fetch lkc-999 --source cluster-health --since 1h", fixture: "logs/fetch/empty.golden"},

		// Environment and organization scope
		{args: "logs fetch --source cluster-health --environment env-596 --since 24h", fixture: "logs/fetch/environment-scope.golden"},
		{args: "logs fetch --source cluster-health --since 24h", fixture: "logs/fetch/organization-scope.golden"},
		{args: "logs fetch --source cluster-health --environment env-596 --since 24h --tail 1", fixture: "logs/fetch/environment-scope-tail.golden"},

		// Validation errors
		{args: "logs fetch lkc-123 --since 1h", fixture: "logs/fetch/missing-source.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health", fixture: "logs/fetch/missing-window.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h --from 2026-10-14T00:00:00Z", fixture: "logs/fetch/since-with-from.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 30d", fixture: "logs/fetch/since-too-long.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --from 2026-10-14T00:00:00 ", fixture: "logs/fetch/from-bad-format.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h --tail 5000", fixture: "logs/fetch/tail-too-large.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source no-such-source --since 1h", fixture: "logs/fetch/unknown-source.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h --search severity=ERROR", fixture: "logs/fetch/unknown-search-field.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h --search level=DEBUG", fixture: "logs/fetch/invalid-search-value.golden", exitCode: 1},
		{args: "logs fetch lkc-123 --source cluster-health --since 1h --search level", fixture: "logs/fetch/malformed-search.golden", exitCode: 1},
		{args: "logs fetch lcc-123 --source cluster-health --since 1h", fixture: "logs/fetch/wrong-id-prefix.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

// TestLogsFetchNext walks one query through its pages: the first page stores a cursor, --next continues it, and the
// last page leaves nothing to continue.
func (s *CLITestSuite) TestLogsFetchNext() {
	const query = "logs fetch lkc-777 --source cluster-health --since 24h"

	s.runIntegrationTest(CLITest{login: "cloud", args: query, fixture: "logs/fetch/next-page-1.golden"})
	s.runIntegrationTest(CLITest{workflow: true, args: query + " --next", fixture: "logs/fetch/next-page-2.golden"})
	s.runIntegrationTest(CLITest{workflow: true, args: query + " --next", fixture: "logs/fetch/next-no-more.golden"})

	// A different query does not match the stored one.
	s.runIntegrationTest(CLITest{workflow: true, args: "logs fetch lkc-777 --source cluster-health --since 1h --next", fixture: "logs/fetch/next-no-stored-query.golden", exitCode: 1})
}

func (s *CLITestSuite) TestLogsFetchNextWithTail() {
	s.runIntegrationTest(CLITest{login: "cloud", args: "logs fetch lkc-777 --source cluster-health --since 24h --next --tail 1", contains: "[next tail] were all set", exitCode: 1})
}

func (s *CLITestSuite) TestLogsFetchTailDropsCursor() {
	s.runIntegrationTest(CLITest{login: "cloud", args: "logs fetch lkc-777 --source cluster-health --since 24h --tail 1", fixture: "logs/fetch/next-tail.golden"})
	s.runIntegrationTest(CLITest{workflow: true, args: "logs fetch lkc-777 --source cluster-health --since 24h --next", fixture: "logs/fetch/next-no-more.golden"})
}

// TestLogsFetchCachesDiscovery checks that a log source is discovered once and then reused, and that a query the
// cached source rejects triggers one refresh before the error is shown.
func (s *CLITestSuite) TestLogsFetchCachesDiscovery() {
	const query = "logs fetch lkc-123 --source cluster-health --since 1h"
	before := testserver.LogSourceRequests
	discoveryRequests := func(t *testing.T, want int) {
		require.Equal(t, want, testserver.LogSourceRequests-before)
	}

	s.runIntegrationTest(CLITest{login: "cloud", args: query, fixture: "logs/fetch/resource.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: query, fixture: "logs/fetch/resource.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: query + " --search level=ERROR", fixture: "logs/fetch/resource-level-error.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: query + " --search severity=ERROR", fixture: "logs/fetch/unknown-search-field.golden", exitCode: 1, wantFunc: func(t *testing.T) { discoveryRequests(t, 2) }})
}

func (s *CLITestSuite) TestLogsFetchOutputFile() {
	path := filepath.Join(s.T().TempDir(), "logs.jsonl")

	s.runIntegrationTest(CLITest{
		login:    "cloud",
		args:     fmt.Sprintf("logs fetch lkc-123 --source cluster-health --since 1h --output-file %s", path),
		contains: fmt.Sprintf("Appended 1 log entries to file: %s", path),
		wantFunc: func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.JSONEq(t, `{"cloud_cluster_id":"lkc-123","event_type":"SERVICE_ISSUE_DETECTED","id":"evt-101","level":"ERROR","message":"Broker lkc-123 unreachable","timestamp":"2026-10-14T14:41:08Z"}`, strings.TrimSpace(string(data)))
		},
	})
}

// logsSearchResponses are the error responses of POST /logs/v1/search that the default mock does not serve.
var logsSearchResponses = map[string]*testserver.LogSourcesResponse{
	"unauthenticated":        {Status: 401, Body: `{"errors":[{"status":"401","detail":"Credentials are required to access this resource."}]}`},
	"feature-not-enabled":    {Status: 403, Body: `{"errors":[{"status":"403","detail":"Feature is not enabled for this organization"}]}`},
	"forbidden":              {Status: 403, Body: `{"errors":[{"status":"403","detail":"User is not authorized to perform this action."}]}`},
	"rate-limited":           {Status: 429, Header: map[string]string{"Retry-After": "12"}, Body: `{"errors":[{"status":"429","detail":"Too many log queries. Please see the documentation for current rate limits."}]}`},
	"rate-limited-no-header": {Status: 429, Body: `{"errors":[{"status":"429","detail":"Too many log queries. Please see the documentation for current rate limits."}]}`},
	"server-error":           {Status: 500, Body: `{"errors":[{"status":"500","detail":"Internal server error"}]}`},
}

// TestLogsFetchResponses covers the error responses of POST /logs/v1/search, including the wait time of a 429.
func (s *CLITestSuite) TestLogsFetchResponses() {
	const query = "logs fetch lkc-123 --source cluster-health --since 1h"
	tests := []struct {
		response string
		fixture  string
	}{
		{"unauthenticated", "logs/fetch/unauthenticated.golden"},
		{"feature-not-enabled", "logs/fetch/feature-not-enabled.golden"},
		{"forbidden", "logs/fetch/forbidden.golden"},
		{"rate-limited", "logs/fetch/rate-limited.golden"},
		{"rate-limited-no-header", "logs/fetch/rate-limited-no-header.golden"},
		{"server-error", "logs/fetch/server-error.golden"},
	}

	defer func() { testserver.LogsSearchOverride = nil }()
	for _, tt := range tests {
		testserver.LogsSearchOverride = logsSearchResponses[tt.response]
		s.runIntegrationTest(CLITest{login: "cloud", args: query, fixture: tt.fixture, exitCode: 1})
	}
}

func (s *CLITestSuite) TestLogsFetchNotLoggedIn() {
	s.runIntegrationTest(CLITest{args: "logs fetch lkc-123 --source cluster-health --since 1h", fixture: "logs/fetch/not-logged-in.golden", exitCode: 1})
}
