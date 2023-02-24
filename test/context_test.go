package test

import (
	"fmt"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
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
	tt := CLITest{fixture: "context/create/0.golden", args: s.contextCreateArgs("0")}
	s.runIntegrationTest(tt)
}

func (s *CLITestSuite) TestContextDelete() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: s.contextCreateArgs("0-prompt")},
		{args: "context delete 0 --force", fixture: "context/delete/0.golden"},
		{args: "context delete 0-prompt", input: "y\n", fixture: "context/delete/0-prompt.golden"},
		{args: "context delete 1 --force", fixture: "context/delete/1.golden", wantErrCode: 1},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestDescribe() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: "context use 0"},
		{fixture: "context/describe/0.golden", args: "context describe"},
		{fixture: "context/describe/1.golden", args: "context describe --api-key"},
		{fixture: "context/describe/2.golden", args: "context describe --username", wantErrCode: 1},
		{fixture: "context/describe/3.golden", args: "context describe --api-key", login: "cloud", wantErrCode: 1},
		{fixture: "context/describe/4.golden", args: "context describe --username", login: "cloud"},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestContextList() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{args: s.contextCreateArgs("1")},
		{fixture: "context/list/0.golden", args: "context list"},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
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

	for _, tt := range tests {
		out := runCommand(s.T(), testBin, env, tt.args, 0, "")
		s.validateTestOutput(tt, s.T(), out)
	}
}

func (s *CLITestSuite) TestContextUpdate() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{fixture: "context/update/0.golden", args: "context update 0 --name 1"},
		{fixture: "context/update/0.golden", args: "context describe 1"},
		{fixture: "context/update/1.golden", args: "context update 1", wantErrCode: 1},
		{args: s.contextCreateArgs("2")},
		{fixture: "context/update/2.golden", args: "context update 1 --name 2", wantErrCode: 1},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestContextUse() {
	resetConfiguration(s.T(), false)

	tests := []CLITest{
		{args: s.contextCreateArgs("0")},
		{fixture: "context/use/0.golden", args: "context describe", wantErrCode: 1},
		{fixture: "context/use/1.golden", args: "context use 0"},
		{fixture: "context/use/2.golden", args: "context describe"},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}
