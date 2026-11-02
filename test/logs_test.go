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
