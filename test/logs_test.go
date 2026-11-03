package test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

func (s *CLITestSuite) TestLogsHelp() {
	tests := []CLITest{
		{args: "logs --help", fixture: "logs/help.golden"},
		{args: "logs source --help", fixture: "logs/source/help.golden"},
		{args: "logs source list --help", fixture: "logs/source/list-help.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

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

// TestLogsSourceListRefreshesDiscovery checks that `source list` always asks the API.
func (s *CLITestSuite) TestLogsSourceListRefreshesDiscovery() {
	before := testserver.LogSourceRequests()
	discoveryRequests := func(t *testing.T, want int64) {
		require.Equal(t, want, testserver.LogSourceRequests()-before)
	}

	s.runIntegrationTest(CLITest{login: "cloud", args: "logs source list", fixture: "logs/source/list.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 1) }})
	s.runIntegrationTest(CLITest{workflow: true, args: "logs source list", fixture: "logs/source/list.golden", wantFunc: func(t *testing.T) { discoveryRequests(t, 2) }})
}

func (s *CLITestSuite) TestLogsSourceListNotLoggedIn() {
	s.runIntegrationTest(CLITest{args: "logs source list", fixture: "logs/source/list-not-logged-in.golden", exitCode: 1})
}

// logSourcesResponses are the responses of GET /logs/v1/sources that the default mock does not serve: sources of other
// shapes, and each error the LLD lists for the endpoint.
var logSourcesResponses = map[string]*testserver.LogSourcesOverride{
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

	defer testserver.SetLogSourcesOverride(nil)
	for _, tt := range tests {
		testserver.SetLogSourcesOverride(logSourcesResponses[tt.response])
		tt.test.login = "cloud"
		s.runIntegrationTest(tt.test)
	}
}
