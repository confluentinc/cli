package test

func (s *CLITestSuite) TestAPIKeyCommands() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --cluster bob", useKafka: "bob", fixture: "apikey_create_1.golden"},
		{args: "api-key list", useKafka: "bob", fixture: "apikey_list_1.golden"},
		{args: "api-key list", useKafka: "abc", fixture: "apikey_list_2.golden"},

		//{args: "kafka cluster list", fixture: "apikey1.golden"},
		//
		//{args: "login --url"+loginURL, env: []string{"XX_CCLOUD_EMAIL=fake@user.com", "XX_CCLOUD_PASSWORD=pass1"}},
		//
		//// create api key for active kafka cluster
		//{args: "kafka cluster use lkc-cool1", fixture: "apikey2.golden"},
		//{args: "api-key list", fixture: "apikey3.golden"},
		//{args: "api-key create --description \"my cool app\"", fixture: "apikey4.golden"},
		//{args: "api-key list", fixture: "apikey5.golden"},

		//// create api key for other kafka cluster
		//{args: "api-key create --description \"my other app\" --cluster lkc-other1", fixture: "apikey6.golden"},
		//{args: "api-key list", fixture: "apikey3.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey7.golden"},
		//
		//// create api key for non-kafka cluster
		//{args: "api-key create --description \"my ksql app\" --cluster lksqlc-ksql1", fixture: "apikey8.golden"},
		//{args: "api-key list", fixture: "apikey3.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey9.golden"},
		//
		//// use an api key for active kafka cluster
		//{args: "api-key use ABCDEF1234", fixture: "apikey10.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//
		//// use an api key for other kafka cluster
		//{args: "api-key use DEFGHI5678 --cluster lkc-other1", fixture: "apikey12.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey13.golden"},
		//
		//// use an api key for non-kafka cluster
		//{args: "api-key use GHIJKL7890 --cluster lksqlc-ksql1", fixture: "apikey14.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey15.golden"},
		//
		//// store an api-key for active kafka cluster
		//{args: "api-key store JKLMNO0987 SECRET1", fixture: "apikey16.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//
		//// store an api-key for other kafka cluster
		//{args: "api-key store MNOPQR6543 SECRET2 --cluster lkc-other1", fixture: "apikey17.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey13.golden"},
		//
		//// store an api-key for non-kafka cluster
		//{args: "api-key store PQRSTU2109 SECRET3 --cluster lksqlc-ksql1", fixture: "apikey18.golden"},
		//{args: "api-key list", fixture: "apikey11.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey15.golden"},
		//
		//// store: error handling
		//{name: "error if storing unknown api key", args: "api-key store UNKNOWN", fixture: "apikey19.golden"},
		//{name: "error if storing api key with existing secret", args: "api-key store EXISTING", fixture: "apikey20.golden"},
		//{name: "succeed if forced to overwrite existing secret", args: "api-key store EXISTING -f", fixture: "apikey21.golden"},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.login = "default"
		tt.workflow = true
		s.runTest(tt, serve(s.T()).URL, serveKafkaAPI(s.T()).URL)
	}
}
