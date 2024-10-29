package test

func (s *CLITestSuite) TestCustomCodeLogging() {
	tests := []CLITest{
		// create
		{args: `ccl custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --destination-topic topic-123 --destination-cluster-id cluster-123`, fixture: "ccl/custom-code-logging/create.golden"},
		//create - missing cloud
		{args: `ccl custom-code-logging create --region us-west-2 --environment env-000000 --destination-kafka --destination-topic topic-123 --destination-cluster-id cluster-123`, fixture: "ccl/custom-code-logging/create-no-cloud.golden", exitCode: 1},
		// create - missing region
		{args: `ccl custom-code-logging create --cloud AWS --environment env-000000 --destination-kafka --destination-topic topic-123 --destination-cluster-id cluster-123`, fixture: "ccl/custom-code-logging/create-no-region.golden", exitCode: 1},
		// create - missing environment
		{args: `ccl custom-code-logging create --cloud AWS --region us-west-2 --destination-kafka --destination-topic topic-123 --destination-cluster-id cluster-123`, fixture: "ccl/custom-code-logging/create-no-env.golden", exitCode: 1},
		// create - missing destination
		{args: `ccl custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000`, fixture: "ccl/custom-code-logging/create-no-destination.golden", exitCode: 1},
		// create - missing destination topic
		{args: `ccl custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --destination-cluster-id cluster-123`, fixture: "ccl/custom-code-logging/create-no-destinationtopic.golden", exitCode: 1},
		// create - missing destination cluster id
		{args: `ccl custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --destination-topic topic-123`, fixture: "ccl/custom-code-logging/create-no-destinationclusterid.golden", exitCode: 1},

		// list
		{args: "ccl custom-code-logging list --environment env-000000", fixture: "ccl/custom-code-logging/list.golden"},
		// list - no environment
		{args: "ccl custom-code-logging list", fixture: "ccl/custom-code-logging/list-no-environment.golden", exitCode: 1},
		// list - yaml
		{args: "ccl custom-code-logging list --environment env-000000 -o json", fixture: "ccl/custom-code-logging/list-json.golden"},
		// list - json
		{args: "ccl custom-code-logging list --environment env-000000 -o yaml", fixture: "ccl/custom-code-logging/list-yaml.golden"},

		// describe
		{args: "ccl custom-code-logging describe ccl-123456", fixture: "ccl/custom-code-logging/describe.golden"},
		// describe - json
		{args: "ccl custom-code-logging describe ccl-456789 -o json", fixture: "ccl/custom-code-logging/describe-json.golden"},
		// describe - yaml
		{args: "ccl custom-code-logging describe ccl-789012 -o yaml", fixture: "ccl/custom-code-logging/describe-yaml.golden"},

		// delete - force
		{args: "ccl custom-code-logging delete ccl-123456 --force", fixture: "ccl/custom-code-logging/delete.golden"},
		// delete
		{args: "ccl custom-code-logging delete ccl-123456", input: "y\n", fixture: "ccl/custom-code-logging/delete-prompt.golden"},
		// update
		{args: "ccl custom-code-logging update ccl-123456 --destination-topic topic-456 --destination-cluster-id cluster-456 --log-level ERROR", fixture: "ccl/custom-code-logging/update.golden"},
		// update - no updates
		{args: "ccl custom-code-logging update ccl-123456", fixture: "ccl/custom-code-logging/update-no-updates.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
