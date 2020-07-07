package test

func (s *CLITestSuite) TestContextCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "config context current", fixture: "context1.golden"},
		{args: "config context list", fixture: "context2.golden"},
		{args: "init my-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", fixture: "context3.golden"},
		{args: "config context set my-context --kafka-cluster anonymous-id", fixture: "context4.golden"},
		{args: "config context list", fixture: "context5.golden"},
		{args: "config context get my-context", fixture: "context6.golden"},
		{args: "config context get other-context", fixture: "context7.golden", wantErrCode: 1},
		{args: "init other-context --kafka-auth --bootstrap boot-test.com --api-key hi --api-secret @test/fixtures/input/apisecret1.txt", fixture: "context8.golden"},
		{args: "config context list", fixture: "context9.golden"},
		{args: "config context use my-context", fixture: "context10.golden"},
		{args: "config context current", fixture: "context11.golden"},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}
