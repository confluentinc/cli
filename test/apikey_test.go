package test

func (s *CLITestSuite) TestAPIKeyCommands() {
	loginURL := serve(s.T()).URL

	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "api-key create --cluster bob", login: "default", fixture: "apikey1.golden"},
		{args: "api-key list", useKafka: "bob", fixture: "apikey2.golden"},
		{args: "api-key list", useKafka: "abc", fixture: "apikey3.golden"},

		// create api key for active kafka cluster
		{args: "kafka cluster use lkc-cool1", fixture: "empty.golden"},
		{args: "api-key list", fixture: "apikey4.golden"},
		{args: "api-key create --description my-cool-app", fixture: "apikey5.golden"},
		{args: "api-key list", fixture: "apikey6.golden"},

		// create api key for other kafka cluster
		{args: "api-key create --description my-other-app --cluster lkc-other1", fixture: "apikey7.golden"},
		//{args: "api-key list", fixture: "apikey4.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey8.golden"},
		//
		//// create api key for non-kafka cluster
		//{args: "api-key create --description my-ksql-app --cluster lksqlc-ksql1", fixture: "apikey9.golden"},
		//{args: "api-key list", fixture: "apikey4.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey10.golden"},
		//
		//// use an api key for active kafka cluster
		//{args: "api-key use MYKEY2", fixture: "empty.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},

		//// use an api key for other kafka cluster
		//{args: "api-key use DEFGHI5678 --cluster lkc-other1", fixture: "apikey13.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey14.golden"},
		//
		//// use an api key for non-kafka cluster
		//{args: "api-key use GHIJKL7890 --cluster lksqlc-ksql1", fixture: "apikey15.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey16.golden"},
		//
		//// store an api-key for active kafka cluster
		//{args: "api-key store JKLMNO0987 SECRET1", fixture: "apikey17.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},
		//
		//// store an api-key for other kafka cluster
		//{args: "api-key store MNOPQR6543 SECRET2 --cluster lkc-other1", fixture: "apikey18.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},
		//{args: "api-key list --cluster lkc-other1", fixture: "apikey14.golden"},
		//
		//// store an api-key for non-kafka cluster
		//{args: "api-key store PQRSTU2109 SECRET3 --cluster lksqlc-ksql1", fixture: "apikey19.golden"},
		//{args: "api-key list", fixture: "apikey12.golden"},
		//{args: "api-key list --cluster lksqlc-ksql1", fixture: "apikey16.golden"},
		//
		//// store: error handling
		//{name: "error if storing unknown api key", args: "api-key store UNKNOWN", fixture: "apikey20.golden"},
		//{name: "error if storing api key with existing secret", args: "api-key store EXISTING", fixture: "apikey21.golden"},
		//{name: "succeed if forced to overwrite existing secret", args: "api-key store EXISTING -f", fixture: "apikey22.golden"},
	}
	resetConfiguration(s.T())
	for _, tt := range tests {
		if tt.name == "" {
			tt.name = tt.args
		}
		tt.workflow = true
		s.runTest(tt, loginURL, serveKafkaAPI(s.T()).URL)
	}
}
