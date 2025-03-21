package test

import (
	"testing"

	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

func (s *CLITestSuite) TestOrganization() {
	tests := []CLITest{
		{args: "organization describe", fixture: "organization/describe.golden"},
		{args: "organization describe -o json", fixture: "organization/describe-json.golden"},
		{args: "organization list", fixture: "organization/list.golden"},
		{args: "organization list -o json", fixture: "organization/list-json.golden"},
		{args: "organization update --name default-updated", fixture: "organization/update.golden"},
		{args: "organization update --name default-updated -o json", fixture: "organization/update-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	// Test organization not found error
	s.T().Run("organization describe not found", func(t *testing.T) {
		testserver.TestOrgNotFound = true
		defer func() { testserver.TestOrgNotFound = false }()

		test := CLITest{
			args:     "organization describe",
			fixture:  "organization/describe-not-found.golden",
			exitCode: 1,
			login:    "cloud",
		}
		s.runIntegrationTest(test)
	})

	s.T().Run("organization describe not found json", func(t *testing.T) {
		testserver.TestOrgNotFound = true
		defer func() { testserver.TestOrgNotFound = false }()

		test := CLITest{
			args:     "organization describe -o json",
			fixture:  "organization/describe-not-found-json.golden",
			exitCode: 1,
			login:    "cloud",
		}
		s.runIntegrationTest(test)
	})
}
