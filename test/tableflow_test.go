package test

func (s *CLITestSuite) TestTableflowCatalogIntegration() {
	tests := []CLITest{
		{args: "tableflow catalog-integration create my-aws-glue-ci --cluster lkc-123456 --type aws --provider-integration cspi-stgce89r7", fixture: "tableflow/catalog-integration/create-aws-glue.golden"},
		{args: "tableflow catalog-integration create my-snowflake-ci --cluster lkc-123456 --type snowflake --endpoint https://vuser1_polaris.snowflakecomputing.com/ --client-id client-id --client-secret client-secret --warehouse warehouse --allowed-scope allowed-scope", fixture: "tableflow/catalog-integration/create-snowflake.golden"},
		{args: "tableflow catalog-integration create my-catalog-integration --cluster lkc-123456 --type unity --workspace-endpoint https://dbc-1.cloud.databricks.com --catalog-name tableflow-quickstart-catalog --unity-client-id $CLIENT_ID --unity-client-secret $CLIENT_SECRET", fixture: "tableflow/catalog-integration/create-unity.golden"},
		{args: "tableflow catalog-integration delete tci-abc123 tci-def456 --cluster lkc-123456", input: "y\n", fixture: "tableflow/catalog-integration/delete-multiple.golden"},
		{args: "tableflow catalog-integration delete tci-abc123 tci-def456 tci-invalid --cluster lkc-123456", fixture: "tableflow/catalog-integration/delete-invalid.golden", exitCode: 1},
		{args: "tableflow catalog-integration list --cluster lkc-123456", fixture: "tableflow/catalog-integration/list.golden"},
		{args: "tableflow catalog-integration list --cluster lkc-123456 -o json", fixture: "tableflow/catalog-integration/list-json.golden"},
		{args: "tableflow catalog-integration describe tci-abc123 --cluster lkc-123456", fixture: "tableflow/catalog-integration/describe-aws-glue.golden"},
		{args: "tableflow catalog-integration describe tci-abc123 --cluster lkc-123456 -o json", fixture: "tableflow/catalog-integration/describe-aws-glue-json.golden"},
		{args: "tableflow catalog-integration describe tci-def456 --cluster lkc-123456", fixture: "tableflow/catalog-integration/describe-snowflake.golden"},
		{args: "tableflow catalog-integration describe tci-ghi789 --cluster lkc-123456", fixture: "tableflow/catalog-integration/describe-unity.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowCatalogIntegrationUpdate() {
	tests := []CLITest{
		{args: "tableflow catalog-integration update tci-def456 --cluster lkc-123456 --endpoint https://vuser2_polaris.snowflakecomputing.com/ --client-id client-id-2 --client-secret client-secret-2 --warehouse warehouse-2 --allowed-scope allowed-scope-2", fixture: "tableflow/catalog-integration/update-snowflake.golden"},
		{args: "tableflow catalog-integration update tci-abc456 --cluster lkc-123456 --name new-name", fixture: "tableflow/catalog-integration/update-name.golden"},
		{args: "tableflow catalog-integration update tci-def456 --cluster lkc-123456 --name new-name --endpoint https://vuser2_polaris.snowflakecomputing.com/ --client-id client-id-2 --client-secret client-secret-2 --warehouse warehouse-2 --allowed-scope allowed-scope-2", fixture: "tableflow/catalog-integration/update-snowflake-with-name.golden"},
		{args: "tableflow catalog-integration update tci-abc123 --cluster lkc-123456", fixture: "tableflow/catalog-integration/update-fail-no-flags.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowCatalogIntegration_Autocomplete() {
	tests := []CLITest{
		{args: `__complete tableflow catalog-integration describe --cluster lkc-123456 ""`, fixture: "tableflow/catalog-integration/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowTopic() {
	tests := []CLITest{
		{args: "tableflow topic enable topic-byob --cluster lkc-123456 --retention-ms 604800000 --storage-type BYOS --provider-integration cspi-stgce89r7 --bucket-name bucket_1 --record-failure-strategy SKIP", fixture: "tableflow/topic/enable-topic-byob.golden"},
		{args: "tableflow topic enable topic-managed --cluster lkc-123456 --retention-ms 604800000 --storage-type MANAGED --table-formats DELTA", fixture: "tableflow/topic/enable-topic-managed.golden"},
		{args: "tableflow topic enable topic-azure --cluster lkc-123456 --retention-ms 604800000 --storage-type AzureDataLakeStorageGen2 --provider-integration cspi-stgce89r7 --container-name container1 --storage-account-name acc1", fixture: "tableflow/topic/enable-topic-azure.golden"},
		{args: "tableflow topic create topic-byob --cluster lkc-123456 --retention-ms 604800000 --storage-type BYOS --provider-integration cspi-stgce89r7 --bucket-name bucket_1 --record-failure-strategy SKIP", fixture: "tableflow/topic/enable-topic-byob.golden"},
		{args: "tableflow topic create topic-managed --cluster lkc-123456 --retention-ms 604800000 --storage-type MANAGED --table-formats DELTA", fixture: "tableflow/topic/enable-topic-managed.golden"},
		{args: "tableflow topic enable topic-azure --cluster lkc-123456 --retention-ms 604800000 --storage-type AzureDataLakeStorageGen2 --provider-integration cspi-stgce89r7 --container-name container1 --storage-account-name acc1", fixture: "tableflow/topic/enable-topic-azure.golden"},
		{args: "tableflow topic enable topic-managed --cluster lkc-123456 --storage-type MANAGED --error-handling SUSPEND", fixture: "tableflow/topic/enable-topic-managed-error-handling-suspend.golden"},
		{args: "tableflow topic enable topic-managed --cluster lkc-123456 --storage-type MANAGED --error-handling SKIP", fixture: "tableflow/topic/enable-topic-managed-error-handling-skip.golden"},
		{args: "tableflow topic enable topic-managed --cluster lkc-123456 --storage-type MANAGED --error-handling LOG --log-target log_topic", fixture: "tableflow/topic/enable-topic-managed-error-handling-log.golden"},
		{args: "tableflow topic update topic-byob --cluster lkc-123456 --retention-ms 432000000", fixture: "tableflow/topic/update-topic.golden"},
		{args: "tableflow topic update topic-managed --cluster lkc-123456 --error-handling SUSPEND", fixture: "tableflow/topic/update-topic-managed-error-handling-suspend.golden"},
		{args: "tableflow topic update topic-managed --cluster lkc-123456 --error-handling SKIP", fixture: "tableflow/topic/update-topic-managed-error-handling-skip.golden"},
		{args: "tableflow topic update topic-managed --cluster lkc-123456 --error-handling LOG --log-target log_topic", fixture: "tableflow/topic/update-topic-managed-error-handling-log.golden"},
		{args: "tableflow topic update topic-managed --cluster lkc-123456 --log-target log_topic", fixture: "tableflow/topic/update-topic-managed-no-change.golden"},
		{args: "tableflow topic update topic-error-log --cluster lkc-123456 --log-target log_topic", fixture: "tableflow/topic/update-topic-error-log.golden"},
		{args: "tableflow topic disable topic-managed --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/disable-topic.golden"},
		{args: "tableflow topic disable topic-managed topic-byob --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/disable-multiple-topics.golden"},
		{args: "tableflow topic disable topic-azure --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/disable-topic-azure.golden"},
		{args: "tableflow topic delete topic-managed --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/delete-topic.golden"},
		{args: "tableflow topic delete topic-managed topic-byob --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/delete-multiple-topics.golden"},
		{args: "tableflow topic delete topic-azure --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/delete-azure-topic.golden"},
		{args: "tableflow topic delete invalid-topic --cluster lkc-123456", input: "y\n", fixture: "tableflow/topic/delete-topic-invalid-1.golden", exitCode: 1},
		{args: "tableflow topic delete invalid-topic --cluster lkc-invalid", input: "y\n", fixture: "tableflow/topic/delete-topic-invalid-2.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowTopicDescribe() {
	tests := []CLITest{
		{args: "tableflow topic describe topic-byob --cluster lkc-123456", fixture: "tableflow/topic/describe-topic.golden"},
		{args: "tableflow topic describe topic-azure --cluster lkc-123456", fixture: "tableflow/topic/describe-topic-azure.golden"},
		{args: "tableflow topic describe topic-byob --cluster lkc-123456 --output json", fixture: "tableflow/topic/describe-topic-json.golden"},
		{args: "tableflow topic describe topic-azure --cluster lkc-123456 --output json", fixture: "tableflow/topic/describe-topic-azure-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowTopicList() {
	tests := []CLITest{
		{args: "tableflow topic list --cluster lkc-123456", fixture: "tableflow/topic/list-topic.golden"},
		{args: "tableflow topic list --cluster lkc-123456 --output json", fixture: "tableflow/topic/list-topic-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestTableflowTopic_Autocomplete() {
	tests := []CLITest{
		{args: `__complete tableflow topic describe --cluster lkc-123456 ""`, fixture: "tableflow/topic/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
