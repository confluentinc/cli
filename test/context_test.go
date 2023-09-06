package test

import (
	"fmt"

	pauth "github.com/confluentinc/cli/v3/pkg/auth"
)

func (s *CLITestSuite) contextCreateArgs(name string) string {
	var (
		bootstrap = s.TestBackend.GetCloudUrl()
		apiKey    = "test"
		apiSecret = "@test/fixtures/input/context/api-secret.txt"
	)

	return fmt.Sprintf("context create %s --bootstrap %s --api-key %s --api-secret %s", name, bootstrap, apiKey, apiSecret)
}

func (s *CLITestSuite) TestContextCreate() {
	resetConfiguration(s.T(), false)
	test := CLITest{fixture: "context/create/0.golden", args: s.contextCreateArgs("0")}
	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestContextDelete() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: s.contextCreateArgs("2")},
		{args: s.contextCreateArgs("3")},
		{args: s.contextCreateArgs("4")},
		{args: s.contextCreateArgs("5")},
		{args: s.contextCreateArgs("0-prompt")},
		{args: "context delete 0 --force", fixture: "context/delete/success.golden"},
		{args: "context delete 0-prompt", input: "y\n", fixture: "context/delete/success-prompt.golden"},
		{args: "context delete 1", fixture: "context/delete/fail.golden", exitCode: 1},
		{args: "context delete 2 3 7 8 9", fixture: "context/delete/multiple-fail.golden", exitCode: 1},
		{args: "context delete 2 3 4", input: "n\n", fixture: "context/delete/multiple-refuse.golden"},
		{args: "context delete 2 3 4", input: "y\n", fixture: "context/delete/multiple-success.golden"},
		{args: "context delete 5 5 5", input: "y\n", fixture: "context/delete/check-success-operation-fail.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestContextDescribe() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: "context use 0"},
		{args: "context describe", fixture: "context/describe/0.golden"},
		{args: "context describe --api-key", fixture: "context/describe/1.golden"},
		{args: "context describe --username", fixture: "context/describe/2.golden", exitCode: 1},
		{args: "context describe --api-key", login: "cloud", fixture: "context/describe/3.golden", exitCode: 1},
		{args: "context describe --username", login: "cloud", fixture: "context/describe/4.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestContextList() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: s.contextCreateArgs("1")},
		{fixture: "context/list/0.golden", args: "context list"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestContextList_CloudAndOnPrem() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: "login --url " + s.TestBackend.GetCloudUrl()},
		{args: "login --url " + s.TestBackend.GetMdsUrl()},
		{fixture: "context/list/1.golden", args: "context list -o yaml", regex: true},
	}

	env := []string{
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformUsername, "on-prem@example.com"),
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformPassword, "password"),
		fmt.Sprintf("%s=%s", pauth.ConfluentCloudEmail, "cloud@example.com"),
		fmt.Sprintf("%s=%s", pauth.ConfluentCloudPassword, "password"),
	}

	for _, test := range tests {
		out := runCommand(s.T(), testBin, env, test.args, 0, "")
		s.validateTestOutput(test, s.T(), out)
	}
}

func (s *CLITestSuite) TestContextUpdate() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{fixture: "context/update/0.golden", args: "context update 0 --name 1"},
		{fixture: "context/update/0.golden", args: "context describe 1"},
		{fixture: "context/update/1.golden", args: "context update 1", exitCode: 1},
		{args: s.contextCreateArgs("2")},
		{fixture: "context/update/2.golden", args: "context update 1 --name 2", exitCode: 1},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestContextUse() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{fixture: "context/use/0.golden", args: "context describe", exitCode: 1},
		{fixture: "context/use/1.golden", args: "context use 0"},
		{fixture: "context/use/2.golden", args: "context describe"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestContextAutocomplete() {
	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: `__complete context describe ""`, fixture: "context/describe/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
