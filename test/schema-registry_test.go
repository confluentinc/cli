package test

func (s *CLITestSuite) TestSrCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// Show what commands are available
		{args: "schema-registry --help", fixture: "schema-registry-help.golden"},
		{args: "schema-registry cluster --help", fixture: "schema-registry-cluster-help.golden"},
		{args: "schema-registry schema --help", fixture: "schema-registry-schema-help.golden"},
		{args: "schema-registry subject --help", fixture: "schema-registry-subject-help.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us", fixture: "schema-registry-enable.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o json", fixture: "schema-registry-enable-json.golden"},
		{args: "schema-registry cluster enable --cloud gcp --geo us -o yaml", fixture: "schema-registry-enable-yaml.golden"},
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
