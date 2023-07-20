package test

func (s *CLITestSuite) TestParseFlagsIntoContext() {
	tests := []CLITest{
		{args: "kafka cluster describe abc --environment not-595", fixture: "dynamic-context/cluster-describe-env-name.golden"},
		{args: "kafka cluster describe abc --environment other", fixture: "dynamic-context/cluster-describe-env-name.golden"},
		{args: "kafka cluster describe abc --environment dne", fixture: "dynamic-context/environment-not-found.golden", exitCode: 1},
		{args: "kafka quota list --cluster lkc-123", fixture: "dynamic-context/kafka-quota-list-using-cluster-name.golden"},
		{args: "kafka quota list --cluster abc", fixture: "dynamic-context/kafka-quota-list-using-cluster-name.golden"},
		{args: "kafka quota list --cluster dne", fixture: "dynamic-context/kafka-cluster-not-found.golden", exitCode: 1},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
