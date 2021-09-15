package test

import (
	"fmt"

	pauth "github.com/confluentinc/cli/internal/pkg/auth"
)

func (s *CLITestSuite) TestConfig() {
	tests := []CLITest{
		{args: "config context current", fixture: "config/1.golden"},
		{args: "config context current --username", fixture: "config/15.golden"},
		{args: "config context list", fixture: "config/2.golden"},
		{args: fmt.Sprintf("init my-context --kafka-auth --bootstrap %s --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", s.TestBackend.GetCloudUrl()), fixture: "config/3.golden", login: "default"},
		{args: "config context set my-context --kafka-cluster anonymous-id", fixture: "config/4.golden"},
		{args: "config context list", fixture: "config/5.golden"},
		{args: "config context get my-context", fixture: "config/6.golden"},
		{args: "config context get other-context", fixture: "config/7.golden", wantErrCode: 1},
		{args: fmt.Sprintf("init other-context --kafka-auth --bootstrap %s --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", s.TestBackend.GetCloudUrl()), fixture: "config/8.golden"},
		{args: "config context list", fixture: "config/9.golden"},
		{args: "config context use my-context", fixture: "config/10.golden"},
		{args: "config context current", fixture: "config/11.golden"},
		{args: "config context current --username", fixture: "config/12.golden"},
		{args: "config context current", login: "default", fixture: "config/13.golden"},
		{args: "config context current --username", login: "default", fixture: "config/14.golden"},
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConfig_CloudAndOnPrem() {
	tests := []CLITest{
		{fixture: "config/16.golden", args: "login --url " + s.TestBackend.GetCloudUrl()},
		{fixture: "config/16.golden", args: "login --url " + s.TestBackend.GetMdsUrl()},
		{fixture: "config/17.golden", args: "config context list -o yaml", regex: true},
	}

	env := []string{
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformUsername, "on-prem@example.com"),
		fmt.Sprintf("%s=%s", pauth.ConfluentPlatformPassword, "password"),
		fmt.Sprintf("%s=%s", pauth.ConfluentCloudEmail, "cloud@example.com"),
		fmt.Sprintf("%s=%s", pauth.ConfluentCloudPassword, "password"),
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		out := runCommand(s.T(), testBin, env, tt.args, 0)
		s.validateTestOutput(tt, s.T(), out)
	}
}
