package test

import "fmt"

func (s *CLITestSuite) TestKafka() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{args: "kafka cluster --help", fixture: "kafka/kafka-cluster-help.golden"},
		{args: "environment use a-595", fixture: "kafka/0.golden"},
		{args: "kafka cluster list", fixture: "kafka/6.golden"},
		{args: "kafka cluster list -o json", fixture: "kafka/7.golden"},
		{args: "kafka cluster list -o yaml", fixture: "kafka/8.golden"},

		{args: "kafka cluster create", fixture: "kafka/1.golden", wantErrCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone", fixture: "kafka/2.golden"},
		{args: "kafka cluster create my-failed-cluster --cloud oops --region us-east1 --availability single-zone", fixture: "kafka/kafka-cloud-provider-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region oops --availability single-zone", fixture: "kafka/kafka-cloud-region-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type oops", fixture: "kafka/kafka-create-type-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type dedicated --cku 0", fixture: "kafka/kafka-cku-error.golden", wantErrCode: 1},
		{args: "kafka cluster create my-dedicated-cluster --cloud aws --region us-east-1 --type dedicated --cku 1", fixture: "kafka/22.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o json", fixture: "kafka/23.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o yaml", fixture: "kafka/24.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability oops-zone", fixture: "kafka/kafka-availability-zone-error.golden", wantErrCode: 1},

		{args: "kafka cluster update lkc-update ", fixture: "kafka/kafka-create-flag-error.golden", wantErrCode: 1},
		{args: "kafka cluster update lkc-update --name lkc-update-name", fixture: "kafka/26.golden"},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o json", fixture: "kafka/28.golden"},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o yaml", fixture: "kafka/29.golden"},
		{args: "kafka cluster update lkc-update-dedicated --name lkc-update-dedicated-name --cku 2", fixture: "kafka/27.golden"},
		{args: "kafka cluster update lkc-update-dedicated --cku 2", fixture: "kafka/39.golden"},
		{args: "kafka cluster update lkc-update --cku 2", fixture: "kafka/kafka-cluster-expansion-error.golden", wantErrCode: 1},

		{args: "kafka cluster delete", fixture: "kafka/3.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-unknown", fixture: "kafka/kafka-delete-unknown-error.golden", wantErrCode: 1},
		{args: "kafka cluster delete lkc-def973", fixture: "kafka/5.golden"},

		{args: "kafka cluster use a-595", fixture: "kafka/40.golden"},

		{args: "kafka region list", fixture: "kafka/14.golden"},
		{args: "kafka region list -o json", fixture: "kafka/15.golden"},
		{args: "kafka region list -o json", fixture: "kafka/16.golden"},
		{args: "kafka region list --cloud gcp", fixture: "kafka/9.golden"},
		{args: "kafka region list --cloud aws", fixture: "kafka/10.golden"},
		{args: "kafka region list --cloud azure", fixture: "kafka/11.golden"},

		{args: "kafka cluster describe lkc-describe", fixture: "kafka/17.golden"},
		{args: "kafka cluster describe lkc-describe -o json", fixture: "kafka/18.golden"},
		{args: "kafka cluster describe lkc-describe -o yaml", fixture: "kafka/19.golden"},

		{args: "kafka cluster describe lkc-describe-dedicated", fixture: "kafka/30.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated -o json", fixture: "kafka/31.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated -o yaml", fixture: "kafka/32.golden"},

		{args: "kafka cluster describe lkc-describe-dedicated-pending", fixture: "kafka/33.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated-pending -o json", fixture: "kafka/34.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated-pending -o yaml", fixture: "kafka/35.golden"},

		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption", fixture: "kafka/36.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o json", fixture: "kafka/37.golden"},
		{args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o yaml", fixture: "kafka/38.golden"},

		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka/kafka-acls-list.golden"},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBED --topic 'test-topic'", fixture: "kafka/kafka-acls-invalid-operation.golden", wantErrCode: 1},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
	}

	resetConfiguration(s.T(), "ccloud")
	kafkaURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaURL).URL

	for _, tt := range tests {
		tt.login = "default"
		tt.workflow = true
		s.runCcloudTest(tt, loginURL)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicList() {
	kafkaRestURL := serveKafkaRest(s.T()).URL
	tests := []CLITest{
		// Test correct usage
		{args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), fixture: "kafka/confluent/topic/list.golden"},
		// Output should format correctly depending on format argument.
		{args: fmt.Sprintf("kafka topic list --url %s -o human", kafkaRestURL), fixture: "kafka/confluent/topic/list.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o yaml", kafkaRestURL), fixture: "kafka/confluent/topic/list-yaml.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o json", kafkaRestURL), fixture: "kafka/confluent/topic/list-json.golden"},
		// Invalid format string should throw error
		{args: fmt.Sprintf("kafka topic list --url %s -o hello", kafkaRestURL), fixture: "kafka/confluent/topic/list-output-error.golden", wantErrCode: 1, name: "invalid format string should throw error"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest, "")
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicCreate() {
	kafkaRestURL := serveKafkaRest(s.T()).URL
	tests := []CLITest{
		// Test correct usage
		{args: fmt.Sprintf("kafka topic create topic-X --url %s", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-success.golden", wantErrCode: 0, name: "correct URL with default params should create successfully"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions 7 --replication-factor 1 --config retention.ms=100000,compression.type=gzip", kafkaRestURL),
			fixture: "kafka/confluent/topic/create-topic-success.golden", wantErrCode: 0, name: "correct URL with valid optional params should create successfully"},
		// Errors: Does not conform to command specification
		{args: fmt.Sprintf("kafka topic create --url %s", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", wantErrCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions as", kafkaRestURL), contains: "Error: invalid argument \"as\" for \"--partitions\" flag: strconv.ParseInt: parsing \"as\": invalid syntax", wantErrCode: 1, name: "argument type not match should return error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor as", kafkaRestURL), contains: "Error: invalid argument \"as\" for \"--replication-factor\" flag: strconv.ParseInt: parsing \"as\": invalid syntax", wantErrCode: 1, name: "argument type not match should return error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "invalid config should return error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=1,compression", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "invalid config should return error"},
		// Errors: Error in topic creation from invalid argument (from server)
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions -10", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-argument-server-error.golden", wantErrCode: 1, name: "invalid num. partitions should lead to error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor 4", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-argument-server-error.golden", wantErrCode: 1, name: "invalid replication-factor should lead to error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=a", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-argument-server-error.golden", wantErrCode: 1, name: "invalid config value should lead to error"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config asdf=asdf", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-argument-server-error.golden", wantErrCode: 1, name: "invalid config key should lead to error"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest, "")
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicDelete() {
	kafkaRestURL := serveKafkaRest(s.T()).URL
	tests := []CLITest{
		// Test correct usage
		{args: fmt.Sprintf("kafka topic delete topic-exist --url %s", kafkaRestURL), fixture: "kafka/confluent/topic/delete-topic-success.golden", wantErrCode: 0, name: "deleting existing topic with correct url should delete successfully"},
		{args: fmt.Sprintf("kafka topic delete topic-not-exist --url %s", kafkaRestURL), fixture: "kafka/confluent/topic/delete-topic-not-exist-failure.golden", wantErrCode: 1, name: "deleting a non-existent topic should fail"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest, "")
	}
}
