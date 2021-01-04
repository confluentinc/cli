package test

func (s *CLITestSuite) TestCCloudConfig() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "config context current", fixture: "config/1.golden"},
		{args: "config context current --username", fixture: "config/15.golden"},
		{args: "config context list", fixture: "config/2.golden"},
		{args: "init my-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", fixture: "config/3.golden"},
		{args: "config context set my-context --kafka-cluster anonymous-id", fixture: "config/4.golden"},
		{args: "config context list", fixture: "config/5.golden"},
		{args: "config context get my-context", fixture: "config/6.golden"},
		{args: "config context get other-context", fixture: "config/7.golden", wantErrCode: 1},
		{args: "init other-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", fixture: "config/8.golden"},
		{args: "config context list", fixture: "config/9.golden"},
		{args: "config context use my-context", fixture: "config/10.golden"},
		{args: "config context current", fixture: "config/11.golden"},
		{args: "config context current --username", fixture: "config/12.golden"},
		{args: "config context current", login: "default", fixture: "config/13.golden"},
		{args: "config context current --username", login: "default", fixture: "config/14.golden"},
	}

	resetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConfluentConfig() {
	tests := []CLITest{
		{args: "config context current", fixture: "config/16.golden"},
		{args: "config context current --username", fixture: "config/17.golden"},
		{args: "config context list", login: "default", fixture: "config/18.golden"},
		{args: "config context current", login: "default", fixture: "config/19.golden"},
		{args: "config context current --username", login: "default", fixture: "config/20.golden"},
	}

	resetConfiguration(s.T(), "confluent")

	for _, tt := range tests {
		tt.workflow = true
		s.runConfluentTest(tt)
	}
}
