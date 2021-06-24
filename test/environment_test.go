package test

func (s *CLITestSuite) TestEnvironment() {
	tests := []CLITest{
		// only login at the begginning so active env is not reset
		// tt.workflow=true so login is not reset
		{Args: "environment list", Fixture: "environment/1.golden", Login: "default"},
		{Args: "environment use not-595", Fixture: "environment/2.golden"},
		{Args: "environment update not-595 --name new-other-name", Fixture: "environment/3.golden"},
		{Args: "environment list", Fixture: "environment/4.golden"},
		{Args: "environment list -o json", Fixture: "environment/5.golden"},
		{Args: "environment list -o yaml", Fixture: "environment/6.golden"},
		{Args: "environment use non-existent-id", Fixture: "environment/7.golden", WantErrCode: 1},
		{Args: "environment create saucayyy", Fixture: "environment/8.golden"},
		{Args: "environment create saucayyy -o json", Fixture: "environment/9.golden"},
		{Args: "environment create saucayyy -o yaml", Fixture: "environment/10.golden"},
		{Args: "environment delete not-595", Fixture: "environment/11.golden"},
	}

	ResetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.Workflow = true
		s.RunCcloudTest(tt)
	}
}
