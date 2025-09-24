package test

func (s *CLITestSuite) TestUsmKafka() {
	tests := []CLITest{
		{args: "usm kafka register 4k0R9d1GTS5tI9f4Y2xZ0Q --name my-kafka-cluster --cloud aws --region us-east-1", fixture: "unified-stream-manager/kafka/create.golden"},
		{args: "usm kafka register 4k0R9d1GTS5tI9f4Y2xZ0Q --name my-kafka-cluster --cloud aws --region us-east-1 -o json", fixture: "unified-stream-manager/kafka/create-json.golden"},
		{args: "usm kafka register 4k0R9d1GTS5tI9f4Y2xZ0Q --name my-kafka-cluster --cloud aws", fixture: "unified-stream-manager/kafka/create-fail-missing-flags.golden", exitCode: 1},
		{args: "usm kafka deregister usmkc-abc123 usmkc-def456", input: "y\n", fixture: "unified-stream-manager/kafka/deregister-multiple.golden"},
		{args: "usm kafka deregister usmkc-invalid", input: "y\n", fixture: "unified-stream-manager/kafka/deregister-invalid.golden", exitCode: 1},
		{args: "usm kafka list", fixture: "unified-stream-manager/kafka/list.golden"},
		{args: "usm kafka list -o json", fixture: "unified-stream-manager/kafka/list-json.golden"},
		{args: "usm kafka describe usmkc-abc123", fixture: "unified-stream-manager/kafka/describe.golden"},
		{args: "usm kafka describe usmkc-invalid", fixture: "unified-stream-manager/kafka/describe-invalid.golden", exitCode: 1},
		// TODO: Uncomment when backend returns resource IDs with the right prefix
		//{args: "usm kafka deregister usm-abc123", fixture: "unified-stream-manager/kafka/deregister-invalid-prefix.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmKafka_Autocomplete() {
	tests := []CLITest{
		{args: `__complete usm kafka describe ""`, fixture: "unified-stream-manager/kafka/describe-autocomplete.golden"},
		{args: `__complete usm kafka deregister ""`, fixture: "unified-stream-manager/kafka/deregister-autocomplete.golden"},
		{args: `__complete usm kafka deregister usmkc-abc123 ""`, fixture: "unified-stream-manager/kafka/register-autocomplete-multiple.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnect() {
	tests := []CLITest{
		{args: "usm connect register connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --cloud aws --region us-east-1", fixture: "unified-stream-manager/connect/create.golden"},
		{args: "usm connect register connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --cloud aws --region us-east-1 -o json", fixture: "unified-stream-manager/connect/create-json.golden"},
		{args: "usm connect register connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q", fixture: "unified-stream-manager/connect/create-default-cloud-region.golden"},
		{args: "usm connect register connect-group-xyz123 --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --cloud aws", fixture: "unified-stream-manager/connect/create-fail-paired-flag-missing.golden", exitCode: 1},
		{args: "usm connect deregister usmcc-abc123 usmcc-def456", input: "y\n", fixture: "unified-stream-manager/connect/deregister-multiple.golden"},
		{args: "usm connect deregister usmcc-invalid", input: "y\n", fixture: "unified-stream-manager/connect/deregister-invalid.golden", exitCode: 1},
		{args: "usm connect deregister usm-abc123", fixture: "unified-stream-manager/connect/deregister-invalid-prefix.golden", exitCode: 1},
		{args: "usm connect list", fixture: "unified-stream-manager/connect/list.golden"},
		{args: "usm connect list -o json", fixture: "unified-stream-manager/connect/list-json.golden"},
		{args: "usm connect describe usmcc-abc123", fixture: "unified-stream-manager/connect/describe.golden"},
		{args: "usm connect describe usmcc-invalid", fixture: "unified-stream-manager/connect/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnect_Autocomplete() {
	tests := []CLITest{
		{args: `__complete usm connect describe ""`, fixture: "unified-stream-manager/connect/describe-autocomplete.golden"},
		{args: `__complete usm connect deregister ""`, fixture: "unified-stream-manager/connect/deregister-autocomplete.golden"},
		{args: `__complete usm connect deregister usmcc-abc123 ""`, fixture: "unified-stream-manager/connect/register-autocomplete-multiple.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
