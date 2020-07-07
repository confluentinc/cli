package test

func (s *CLITestSuite) TestServiceAccountCommands() {
	tests := []CLITest{
		{args: "service-account list", fixture: "service-account1.golden", wantErrCode: 0},
		{args: "service-account list -o json", fixture: "service-account2.golden", wantErrCode: 0},
		{args: "service-account list -o yaml", fixture: "service-account3.golden", wantErrCode: 0},
		{args: "service-account create human-service --description human-output", fixture: "service-account4.golden", wantErrCode: 0},
		{args: "service-account create json-service --description json-output -o json", fixture: "service-account5.golden", wantErrCode: 0},
		{args: "service-account create yaml-service --description yaml-output -o yaml", fixture: "service-account6.golden", wantErrCode: 0},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}
