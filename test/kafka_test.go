package test

func (s *CLITestSuite) TestKafkaCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "kafka cluster --help", fixture: "kafka-cluster-help.golden"},
		{args: "environment use a-595", fixture: "kafka0.golden", wantErrCode: 0},
		{args: "kafka cluster list", fixture: "kafka6.golden", wantErrCode: 0},
		{args: "kafka cluster list -o json", fixture: "kafka7.golden", wantErrCode: 0},
		{args: "kafka cluster list -o yaml", fixture: "kafka8.golden", wantErrCode: 0},
		{args: "kafka cluster create", fixture: "kafka1.golden", wantErrCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1", fixture: "kafka2.golden", wantErrCode: 0},
		{args: "kafka cluster create my-failed-cluster --cloud oops --region us-east1", fixture: "kafka12.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region oops", fixture: "kafka13.golden", wantErrCode: 1},
		{args: "kafka cluster delete", fixture: "kafka3.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-unknown", fixture: "kafka4.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-def973", fixture: "kafka5.golden", wantErrCode: 0},
		{args: "kafka region list", fixture: "kafka14.golden", wantErrCode: 0},
		{args: "kafka region list --cloud gcp", fixture: "kafka9.golden", wantErrCode: 0},
		{args: "kafka region list --cloud aws", fixture: "kafka10.golden", wantErrCode: 0},
		{args: "kafka region list --cloud azure", fixture: "kafka11.golden", wantErrCode: 0},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL)
	}
}
