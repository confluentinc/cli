package test

func (s *CLITestSuite) TestConnectCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "connector --help", fixture: "connector-help.golden"},
		{args: "connector-catalog --help", fixture: "connector-catalog-help.golden"},
		{args: "connector create --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connector-create.golden", wantErrCode: 0},
		{args: "connector list --cluster lkc-123", fixture: "connector-list.golden", wantErrCode: 0},
		{args: "connector describe lcc-123 --cluster lkc-123", fixture: "connector-describe.golden", wantErrCode: 0},
		{args: "connector pause lcc-123 --cluster lkc-123", fixture: "connector-pause.golden", wantErrCode: 0},
		{args: "connector resume lcc-123 --cluster lkc-123", fixture: "connector-resume.golden", wantErrCode: 0},
		{args: "connector delete lcc-123 --cluster lkc-123", fixture: "connector-delete.golden", wantErrCode: 0},
		{args: "connector update lcc-123 --cluster lkc-123 --config test/fixtures/input/connector-config.yaml", fixture: "connector-update.golden", wantErrCode: 0},
		{args: "connector-catalog list --cluster lkc-123", fixture: "connector-catalog-list.golden", wantErrCode: 0},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		urlVal := serve(s.T(), kafkaAPIURL)
		s.runCcloudTest(tt, urlVal.URL, kafkaAPIURL)
	}
}
