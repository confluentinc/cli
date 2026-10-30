package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLITestSuite) TestLogsHelp() {
	tests := []CLITest{
		{args: "logs --help", fixture: "logs/help.golden"},
		{args: "logs fetch --help", fixture: "logs/fetch-help.golden"},
		{args: "logs source --help", fixture: "logs/source/help.golden"},
		{args: "logs source list --help", fixture: "logs/source/list-help.golden"},
		{args: "logs source describe --help", fixture: "logs/source/describe-help.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
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

		// Connector logs: the CRN is built from the connector name
		{args: "logs fetch lcc-123 --source connector --cluster lkc-123 --since 1h", fixture: "logs/fetch/connector.golden"},
		{args: "logs fetch --source connector --since 1h", fixture: "logs/fetch/connector-needs-resource-id.golden", exitCode: 1},

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

func (s *CLITestSuite) TestLogsFetchOutputFile() {
	path := filepath.Join(s.T().TempDir(), "logs.jsonl")

	s.runIntegrationTest(CLITest{
		login:    "cloud",
		args:     fmt.Sprintf("logs fetch lkc-123 --source cluster-health --since 1h --output-file %s", path),
		contains: fmt.Sprintf("Appended 1 log entries to file: %s", path),
		wantFunc: func(t *testing.T) {
			data, err := os.ReadFile(path)
			require.NoError(t, err)
			require.JSONEq(t, `{"cluster_id":"lkc-123","event_type":"SERVICE_ISSUE_DETECTED","id":"evt-101","level":"ERROR","message":"Broker lkc-123 unreachable","timestamp":"2026-10-14T14:41:08Z"}`, strings.TrimSpace(string(data)))
		},
	})
}

func (s *CLITestSuite) TestLogsAutocomplete() {
	tests := []CLITest{
		{args: `__complete logs fetch --source ""`, fixture: "logs/fetch-source-autocomplete.golden"},
		{args: `__complete logs source describe ""`, fixture: "logs/source/describe-autocomplete.golden"},
		{args: `__complete logs source describe cluster-health ""`, fixture: "logs/source/describe-autocomplete-second-arg.golden"},
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

func (s *CLITestSuite) TestLogsSourceDescribe() {
	tests := []CLITest{
		{args: "logs source describe cluster-health", fixture: "logs/source/describe.golden"},
		{args: "logs source describe connector", fixture: "logs/source/describe-connector.golden"},
		{args: "logs source describe cluster-health -o json", fixture: "logs/source/describe-json.golden"},
		{args: "logs source describe cluster-health -o yaml", fixture: "logs/source/describe-yaml.golden"},
		{args: "logs source describe no-such-source", fixture: "logs/source/describe-unknown.golden", exitCode: 1},
		{args: "logs source describe", fixture: "logs/source/describe-missing-arg.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
