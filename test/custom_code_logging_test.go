package test

func (s *CLITestSuite) TestCustomCodeLogging() {
	tests := []CLITest{
		// create
		{args: `custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --topic topic-123 --cluster-id cluster-123`, fixture: "custom-code-logging/create.golden"},
		//create - missing cloud
		{args: `custom-code-logging create --region us-west-2 --environment env-000000 --destination-kafka --topic topic-123 --cluster-id cluster-123`, fixture: "custom-code-logging/create-no-cloud.golden", exitCode: 1},
		// create - missing region
		{args: `custom-code-logging create --cloud AWS --environment env-000000 --destination-kafka --topic topic-123 --cluster-id cluster-123`, fixture: "custom-code-logging/create-no-region.golden", exitCode: 1},
		// create - missing environment
		{args: `custom-code-logging create --cloud AWS --region us-west-2 --destination-kafka --topic topic-123 --cluster-id cluster-123`, fixture: "custom-code-logging/create-no-env.golden", exitCode: 1},
		// create - missing destination
		{args: `custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000`, fixture: "custom-code-logging/create-no-destination.golden", exitCode: 1},
		// create - missing destination topic
		{args: `custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --cluster-id cluster-123`, fixture: "custom-code-logging/create-no-topic.golden", exitCode: 1},
		// create - missing destination cluster id
		{args: `custom-code-logging create --cloud AWS --region us-west-2 --environment env-000000 --destination-kafka --topic topic-123`, fixture: "custom-code-logging/create-no-clusterid.golden", exitCode: 1},

		// list
		{args: "custom-code-logging list --environment env-000000", fixture: "custom-code-logging/list.golden"},
		// list - no environment
		{args: "custom-code-logging list", fixture: "custom-code-logging/list-no-environment.golden", exitCode: 1},
		// list - yaml
		{args: "custom-code-logging list --environment env-000000 -o json", fixture: "custom-code-logging/list-json.golden"},
		// list - json
		{args: "custom-code-logging list --environment env-000000 -o yaml", fixture: "custom-code-logging/list-yaml.golden"},

		// describe
		{args: "custom-code-logging describe ccl-123456", fixture: "custom-code-logging/describe.golden"},
		// describe - json
		{args: "custom-code-logging describe ccl-456789 -o json", fixture: "custom-code-logging/describe-json.golden"},
		// describe - yaml
		{args: "custom-code-logging describe ccl-789012 -o yaml", fixture: "custom-code-logging/describe-yaml.golden"},

		// delete - force
		{args: "custom-code-logging delete ccl-123456 --force", fixture: "custom-code-logging/delete.golden"},
		// delete
		{args: "custom-code-logging delete ccl-123456", input: "y\n", fixture: "custom-code-logging/delete-prompt.golden"},
		// update
		{args: "custom-code-logging update ccl-123456 --topic topic-456 --cluster-id cluster-456 --log-level ERROR", fixture: "custom-code-logging/update.golden"},
		// update - no updates
		{args: "custom-code-logging update ccl-123456", fixture: "custom-code-logging/update-no-updates.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
