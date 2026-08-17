package test

func (s *CLITestSuite) TestUsmConnectClusterCreate() {
	tests := []CLITest{
		{args: "usm connect-cluster create test-name --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q", fixture: "usm/connect-cluster/create/create.golden"},
		{args: "usm connect-cluster create test-name --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --region us-east-1", fixture: "usm/connect-cluster/create/create-region.golden"},
		{args: "usm connect-cluster create test-name --kafka-cluster lkc-abc123", fixture: "usm/connect-cluster/create/create-cloud-kafka.golden"},
		{args: "usm connect-cluster create test-name --kafka-cluster lkc-abc123 -o json", fixture: "usm/connect-cluster/create/create-cloud-kafka-json.golden"},
		{args: "usm connect-cluster create test-name", fixture: "usm/connect-cluster/create/create-fail-missing-kafka-flag.golden", exitCode: 1},
		{args: "usm connect-cluster create test-name --confluent-platform-kafka-cluster 4k0R9d1GTS5tI9f4Y2xZ0Q --kafka-cluster lkc-abc123", fixture: "usm/connect-cluster/create/create-fail-mutually-exclusive-kafka-flags.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnectClusterDelete() {
	tests := []CLITest{
		{args: "usm connect-cluster delete usmcc-1 --force", fixture: "usm/connect-cluster/delete/delete.golden"},
		{args: "usm connect-cluster delete usmcc-1", input: "y\n", fixture: "usm/connect-cluster/delete/delete-no-force.golden"},
		{args: "usm connect-cluster delete usmcc-1 usmcc-2", input: "y\n", fixture: "usm/connect-cluster/delete/delete-multiple.golden"},
		{args: "usm connect-cluster delete invalid", fixture: "usm/connect-cluster/delete/delete-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnectClusterDescribe() {
	tests := []CLITest{
		{args: "usm connect-cluster describe usmcc-1", fixture: "usm/connect-cluster/describe/describe.golden"},
		{args: "usm connect-cluster describe usmcc-1 -o json", fixture: "usm/connect-cluster/describe/describe-json.golden"},
		{args: "usm connect-cluster describe usmcc-1 -o yaml", fixture: "usm/connect-cluster/describe/describe-yaml.golden"},
		{args: "usm connect-cluster describe invalid", fixture: "usm/connect-cluster/describe/describe-invalid.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnectClusterList() {
	tests := []CLITest{
		{args: "usm connect-cluster list", fixture: "usm/connect-cluster/list/list.golden"},
		{args: "usm connect-cluster list -o json", fixture: "usm/connect-cluster/list/list-json.golden"},
		{args: "usm connect-cluster list -o yaml", fixture: "usm/connect-cluster/list/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestUsmConnectCluster_Autocomplete() {
	tests := []CLITest{
		{args: "__complete usm connect-cluster delete \"\"", fixture: "usm/connect-cluster/delete/delete-autocomplete.golden"},
		{args: "__complete usm connect-cluster describe \"\"", fixture: "usm/connect-cluster/describe/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
