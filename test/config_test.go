package test

func (s *CLITestSuite) TestCCloudConfig() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{Args: "config context current", Fixture: "config/1.golden"},
		{Args: "config context current --username", Fixture: "config/15.golden"},
		{Args: "config context list", Fixture: "config/2.golden"},
		{Args: "init my-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/Fixtures/input/apisecret1.txt", Fixture: "config/3.golden"},
		{Args: "config context set my-context --kafka-cluster anonymous-id", Fixture: "config/4.golden"},
		{Args: "config context list", Fixture: "config/5.golden"},
		{Args: "config context get my-context", Fixture: "config/6.golden"},
		{Args: "config context get other-context", Fixture: "config/7.golden", WantErrCode: 1},
		{Args: "init other-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/Fixtures/input/apisecret1.txt", Fixture: "config/8.golden"},
		{Args: "config context list", Fixture: "config/9.golden"},
		{Args: "config context use my-context", Fixture: "config/10.golden"},
		{Args: "config context current", Fixture: "config/11.golden"},
		{Args: "config context current --username", Fixture: "config/12.golden"},
		{Args: "config context current", Login: "default", Fixture: "config/13.golden"},
		{Args: "config context current --username", Login: "default", Fixture: "config/14.golden"},
	}

	ResetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.Workflow = true
		s.RunCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConfluentConfig() {
	tests := []CLITest{
		{Args: "config context current", Fixture: "config/16.golden"},
		{Args: "config context current --username", Fixture: "config/17.golden"},
		{Args: "config context list", Login: "default", Fixture: "config/18.golden"},
		{Args: "config context current", Login: "default", Fixture: "config/19.golden"},
		{Args: "config context current --username", Login: "default", Fixture: "config/20.golden"},
	}

	ResetConfiguration(s.T(), "confluent")

	for _, tt := range tests {
		tt.Workflow = true
		s.RunConfluentTest(tt)
	}
}
