package test

import (
	"fmt"
	"strings"

	"github.com/confluentinc/bincover"
)

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

		{args: "kafka cluster describe lkc-describe-infinite", fixture: "kafka/41.golden"},
		{args: "kafka cluster describe lkc-describe-infinite -o json", fixture: "kafka/42.golden"},
		{args: "kafka cluster describe lkc-describe-infinite -o yaml", fixture: "kafka/43.golden"},

		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka/kafka-acls-list.golden"},
		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka/rp-kafka-acls-list.golden"},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBED --topic 'test-topic'", fixture: "kafka/kafka-acls-invalid-operation.golden", wantErrCode: 1},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},

		{args: "kafka link list --cluster lkc-links", fixture: "kafka/kafka20.golden", wantErrCode: 0},
		{args: "kafka link list --cluster lkc-links -o json", fixture: "kafka/kafka21.golden", wantErrCode: 0},
		{args: "kafka link list --cluster lkc-links -o yaml", fixture: "kafka/kafka22.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link", fixture: "kafka/kafka23.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link -o json", fixture: "kafka/kafka24.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link -o yaml", fixture: "kafka/kafka25.golden", wantErrCode: 0},

		{args: "kafka topic mirror stop test-topic", login: "default", useKafka: "lkc-topics", authKafka: "true"},
		{args: "kafka topic mirror stop not-found", login: "default", useKafka: "lkc-topics", authKafka: "true", fixture: "kafka/mirror-topic-not-found.golden", wantErrCode: 1},
		{args: "kafka topic mirror bad test-topic", login: "default", useKafka: "lkc-topics", authKafka: "true", fixture: "kafka/mirror-invalid.golden", wantErrCode: 1},

		{args: "kafka topic list --cluster lkc-kafka-api-topics", login: "default", fixture: "kafka/topic-list.golden"},
		{args: "kafka topic list --cluster lkc-topics", fixture: "kafka/rp-topic-list.golden"},
		{args: "kafka topic list", login: "default", useKafka: "lkc-kafka-api-no-topics", fixture: "kafka/topic-list-empty.golden"},
		{args: "kafka topic list", login: "default", useKafka: "lkc-not-ready", fixture: "kafka/cluster-not-ready.golden", wantErrCode: 1},

		{args: "kafka topic create", login: "default", useKafka: "lkc-create-topic", fixture: "kafka/topic-create.golden", wantErrCode: 1},
		{args: "kafka topic create topic1", login: "default", useKafka: "lkc-create-topic-kafka-api", fixture: "kafka/topic-create-success.golden"},
		{args: "kafka topic create topic1", useKafka: "lkc-create-topic", fixture: "kafka/topic-create-success.golden"},
		{args: "kafka topic create topic-exist", login: "default", useKafka: "lkc-create-topic", fixture: "kafka/topic-create-dup-topic.golden", wantErrCode: 1},

		{args: "kafka topic describe", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-describe.golden", wantErrCode: 1},
		{args: "kafka topic describe topic1", login: "default", useKafka: "lkc-describe-topic-kafka-api", fixture: "kafka/topic-describe-success.golden"},
		{args: "kafka topic describe topic-exist", useKafka: "lkc-describe-topic", fixture: "kafka/rp-topic-describe-success.golden"},
		{args: "kafka topic describe topic-exist --output json", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-describe-json-success.golden"},
		{args: "kafka topic describe topic1 --cluster lkc-create-topic-kafka-api", login: "default", fixture: "kafka/topic-describe-not-found.golden", wantErrCode: 1},
		{args: "kafka topic describe topic2", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic2-describe-not-found.golden", wantErrCode: 1},

		{args: "kafka topic delete", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete.golden", wantErrCode: 1},
		{args: "kafka topic delete topic-exist", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete-success.golden"},
		{args: "kafka topic delete topic-exist", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete-success.golden"},
		{args: "kafka topic delete topic1 --cluster lkc-create-topic-kafka-api", login: "default", fixture: "kafka/topic-delete-not-found.golden", wantErrCode: 1},
		{args: "kafka topic delete topic2", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic2-delete-not-found.golden", wantErrCode: 1},

		{args: "kafka topic update topic-exist --config retention.ms=1,compression.type=gzip", login: "default", useKafka: "lkc-describe-topic-kafka-api", fixture: "kafka/topic-update-success.golden"},
		{args: "kafka topic update topic-exist --config retention.ms=1,compression.type=gzip", useKafka: "lkc-describe-topic", fixture: "kafka/topic-update-success.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
	}

	resetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.login = "default"
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicList() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Test correct usage
		{args: fmt.Sprintf("kafka topic list --url %s --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/list.golden"},
		// Test with basic auth input
		{args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("Miles\nTod\n"))}, fixture: "kafka/confluent/topic/list-with-auth.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), login: "default", fixture: "kafka/confluent/topic/list-with-auth-from-login.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s --prompt", kafkaRestURL), login: "default", preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("Miles\nTod\n"))}, fixture: "kafka/confluent/topic/list-with-auth-prompt.golden"},
		// Test with CONFLUENT_REST_URL env var
		{args: "kafka topic list --no-auth", fixture: "kafka/confluent/topic/list.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Test failure when only one of client-cert-path or client-key-path are provided
		{args: "kafka topic list --client-cert-path cert.crt", wantErrCode: 1, fixture: "kafka/confluent/topic/client-cert-flag-error.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		{args: "kafka topic list --client-key-path cert.key", wantErrCode: 1, fixture: "kafka/confluent/topic/client-cert-flag-error.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Output should format correctly depending on format argument.
		{args: fmt.Sprintf("kafka topic list --url %s -o human --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/list.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o yaml --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/list-yaml.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o json --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/list-json.golden"},
		// Invalid format string should throw error
		{args: fmt.Sprintf("kafka topic list --url %s -o hello --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/list-output-error.golden", wantErrCode: 1, name: "invalid format string should throw error"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicCreate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// <topic> errors
		{args: fmt.Sprintf("kafka topic create --url %s --no-auth", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", wantErrCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic create topic-exist --url %s --no-auth", kafkaRestURL), contains: "Error: topic \"topic-exist\" already exists for the Kafka cluster\n\nSuggestions:\n    To list topics for the cluster, use `confluent kafka topic list --url <url>`.", wantErrCode: 1, name: "creating topic with existing topic name should fail"},
		// --partitions errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions -2 --no-auth", kafkaRestURL), contains: "Error: REST request failed: Number of partitions must be larger than 0. (40002)\n", wantErrCode: 1, name: "creating topic with negative partitions name should fail"},
		// --replication-factor errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor 4 --no-auth", kafkaRestURL), contains: "Error: REST request failed: Replication factor: 4 larger than available brokers: 3. (40002)\n", wantErrCode: 1, name: "creating topic with larger replication factor than num. brokers should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor -2 --no-auth", kafkaRestURL), contains: "Error: REST request failed: Replication factor must be larger than 0. (40002)\n", wantErrCode: 1, name: "creating topic with negative replication factor should fail"},
		// --config errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "creating topic with poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=1, --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "creating topic with poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=1,compression --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "creating topic with poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config asdf=1 --no-auth", kafkaRestURL), contains: "Error: REST request failed: Unknown topic config name: asdf (40002)\n", wantErrCode: 1, name: "creating topic with incorrect config name should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=as --no-auth", kafkaRestURL), contains: "Error: REST request failed: Invalid value as for configuration retention.ms: Not a number of type LONG (40002)\n", wantErrCode: 1, name: "creating topic with correct key incorrect config value should fail"},
		// Success
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/create-topic-success.golden", wantErrCode: 0, name: "correct URL with default params (part 6, repl 3, no configs) should create successfully"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions 7 --replication-factor 2 --config retention.ms=100000,compression.type=gzip --no-auth", kafkaRestURL),
			fixture: "kafka/confluent/topic/create-topic-success.golden", wantErrCode: 0, name: "correct URL with valid optional params should create successfully"},
		// --ifnotexists
		{args: fmt.Sprintf("kafka topic create topic-exist --url %s --if-not-exists --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/create-duplicate-topic-ifnotexists-success.golden", wantErrCode: 0, name: "create topic with existing topic name with if-not-exists flag should succeed"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicDelete() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: fmt.Sprintf("kafka topic delete --url %s --no-auth", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", wantErrCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic delete topic-exist --url %s --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/delete-topic-success.golden", wantErrCode: 0, name: "deleting existing topic with correct url should delete successfully"},
		{args: fmt.Sprintf("kafka topic delete topic-not-exist --url %s --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/delete-topic-not-exist-failure.golden", wantErrCode: 1, name: "deleting a non-existent topic should fail"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicUpdate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic name errors
		{args: fmt.Sprintf("kafka topic update --url %s --no-auth", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", wantErrCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic update topic-not-exist --url %s --no-auth", kafkaRestURL), contains: "Error: REST request failed: This server does not host this topic-partition. (40403)\n", wantErrCode: 1, name: "update config of a non-existent topic should fail"},
		// --config errors
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1, --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression --no-auth", kafkaRestURL), contains: "Error: configuration must be in the form of key=value", wantErrCode: 1, name: "poorly formatted config arg should fail"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config asdf=1 --no-auth", kafkaRestURL), contains: "Error: REST request failed: Config asdf cannot be found for TOPIC topic-exist in cluster cluster-1. (404)\n", wantErrCode: 1, name: "incorrect config name should fail"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=as --no-auth", kafkaRestURL), contains: "Error: REST request failed: Invalid config value for resource ConfigResource(type=TOPIC, name='topic-exist'): Invalid value as for configuration retention.ms: Not a number of type LONG (40002)\n", wantErrCode: 1, name: "correct key incorrect config value should fail"},
		// Success cases
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/update-topic-config-success", wantErrCode: 0, name: "valid config updates should succeed with configs printed sorted"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1000,retention.ms=1 --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/update-topic-config-duplicate-success", wantErrCode: 0, name: "valid duplicate config should succeed with the later config value kept"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth -o json", kafkaRestURL), fixture: "kafka/confluent/topic/update-topic-config-success-json.golden", wantErrCode: 0, name: "config updates with json output"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth -o yaml", kafkaRestURL), fixture: "kafka/confluent/topic/update-topic-config-success-yaml.golden", wantErrCode: 0, name: "config updates with yaml output"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicDescribe() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic name errors
		{args: fmt.Sprintf("kafka topic describe --url %s --no-auth", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", wantErrCode: 1, name: "<topic> arg missing should lead to error"},
		{args: fmt.Sprintf("kafka topic describe topic-not-exist --url %s --no-auth", kafkaRestURL), contains: "Error: REST request failed: This server does not host this topic-partition. (40403)\n", wantErrCode: 1, name: "describing a non-existant topic should lead to error"},
		// -o errors
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o asdf --no-auth", kafkaRestURL), contains: "Error: invalid value \"asdf\" for flag `--output`\n\nSuggestions:\n    The possible values for flag `output` are: human, json, yaml.", wantErrCode: 1, name: "bad output format flag should lead to error"},
		// Success cases
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/describe-topic-success.golden", wantErrCode: 0, name: "topic that exists & correct format arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o human --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/describe-topic-success.golden", wantErrCode: 0, name: "topic that exist & human arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o json --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/describe-topic-success-json.golden", wantErrCode: 0, name: "topic that exist & json arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o yaml --no-auth", kafkaRestURL), fixture: "kafka/confluent/topic/describe-topic-success-yaml.golden", wantErrCode: 0, name: "topic that exist & yaml arg should lead to success"},
	}

	for _, clitest := range tests {
		s.runConfluentTest(clitest)
	}
}
