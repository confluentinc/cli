package test

import (
	"fmt"
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestKafka() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		{Args: "kafka cluster --help", Fixture: "kafka/kafka-cluster-help.golden"},
		{Args: "Environment use a-595", Fixture: "kafka/0.golden"},
		{Args: "kafka cluster list", Fixture: "kafka/6.golden"},
		{Args: "kafka cluster list -o json", Fixture: "kafka/7.golden"},
		{Args: "kafka cluster list -o yaml", Fixture: "kafka/8.golden"},

		{Args: "kafka cluster create", Fixture: "kafka/1.golden", WantErrCode: 1},
		{Args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone", Fixture: "kafka/2.golden"},
		{Args: "kafka cluster create my-failed-cluster --cloud oops --region us-east1 --availability single-zone", Fixture: "kafka/kafka-cloud-provider-error.golden", WantErrCode: 1},
		{Args: "kafka cluster create my-failed-cluster --cloud aws --region oops --availability single-zone", Fixture: "kafka/kafka-cloud-region-error.golden", WantErrCode: 1},
		{Args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type oops", Fixture: "kafka/kafka-create-type-error.golden", WantErrCode: 1},
		{Args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type dedicated --cku 0", Fixture: "kafka/kafka-cku-error.golden", WantErrCode: 1},
		{Args: "kafka cluster create my-dedicated-cluster --cloud aws --region us-east-1 --type dedicated --cku 1", Fixture: "kafka/22.golden"},
		{Args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o json", Fixture: "kafka/23.golden"},
		{Args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o yaml", Fixture: "kafka/24.golden"},
		{Args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability oops-zone", Fixture: "kafka/kafka-availability-zone-error.golden", WantErrCode: 1},

		{Args: "kafka cluster update lkc-update ", Fixture: "kafka/kafka-create-flag-error.golden", WantErrCode: 1},
		{Args: "kafka cluster update lkc-update --Name lkc-update-Name", Fixture: "kafka/26.golden"},
		{Args: "kafka cluster update lkc-update --Name lkc-update-Name -o json", Fixture: "kafka/28.golden"},
		{Args: "kafka cluster update lkc-update --Name lkc-update-Name -o yaml", Fixture: "kafka/29.golden"},
		{Args: "kafka cluster update lkc-update-dedicated-expand --Name lkc-update-dedicated-Name --cku 2", Fixture: "kafka/27.golden"},
		{Args: "kafka cluster update lkc-update-dedicated-expand --cku 2", Fixture: "kafka/39.golden"},
		{Args: "kafka cluster update lkc-update --cku 2", Fixture: "kafka/kafka-cluster-expansion-error.golden", WantErrCode: 1},
		{Args: "kafka cluster update lkc-update-dedicated-shrink --Name lkc-update-dedicated-Name --cku 1", Fixture: "kafka/44.golden"},
		{Args: "kafka cluster update lkc-update-dedicated-shrink --cku 1", Fixture: "kafka/45.golden"},
		{Args: "kafka cluster update lkc-update --cku 1", Fixture: "kafka/kafka-cluster-shrink-error.golden", WantErrCode: 1},

		{Args: "kafka cluster delete", Fixture: "kafka/3.golden", WantErrCode: 1},
		{Args: "kafka cluster delete lkc-unknown", Fixture: "kafka/kafka-delete-unknown-error.golden", WantErrCode: 1},
		{Args: "kafka cluster delete lkc-def973", Fixture: "kafka/5.golden"},

		{Args: "kafka cluster use a-595", Fixture: "kafka/40.golden"},

		{Args: "kafka region list", Fixture: "kafka/14.golden"},
		{Args: "kafka region list -o json", Fixture: "kafka/15.golden"},
		{Args: "kafka region list -o json", Fixture: "kafka/16.golden"},
		{Args: "kafka region list --cloud gcp", Fixture: "kafka/9.golden"},
		{Args: "kafka region list --cloud aws", Fixture: "kafka/10.golden"},
		{Args: "kafka region list --cloud azure", Fixture: "kafka/11.golden"},

		{Args: "kafka cluster describe lkc-describe", Fixture: "kafka/17.golden"},
		{Args: "kafka cluster describe lkc-describe -o json", Fixture: "kafka/18.golden"},
		{Args: "kafka cluster describe lkc-describe -o yaml", Fixture: "kafka/19.golden"},

		{Args: "kafka cluster describe lkc-describe-dedicated", Fixture: "kafka/30.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated -o json", Fixture: "kafka/31.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated -o yaml", Fixture: "kafka/32.golden"},

		{Args: "kafka cluster describe lkc-describe-dedicated-pending", Fixture: "kafka/33.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated-pending -o json", Fixture: "kafka/34.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated-pending -o yaml", Fixture: "kafka/35.golden"},

		{Args: "kafka cluster describe lkc-describe-dedicated-with-encryption", Fixture: "kafka/36.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o json", Fixture: "kafka/37.golden"},
		{Args: "kafka cluster describe lkc-describe-dedicated-with-encryption -o yaml", Fixture: "kafka/38.golden"},

		{Args: "kafka cluster describe lkc-describe-infinite", Fixture: "kafka/41.golden"},
		{Args: "kafka cluster describe lkc-describe-infinite -o json", Fixture: "kafka/42.golden"},
		{Args: "kafka cluster describe lkc-describe-infinite -o yaml", Fixture: "kafka/43.golden"},

		{Args: "kafka acl list --cluster lkc-acls-kafka-api", Fixture: "kafka/kafka-acls-list.golden"},
		//{Args: "kafka acl list --cluster lkc-acls", Fixture: "kafka/rp-kafka-acls-list.golden"},
		//{Args: "kafka acl list --cluster lkc-acls -o json", Fixture: "kafka/rp-kafka-acls-list-json.golden"},
		//{Args: "kafka acl list --cluster lkc-acls -o yaml", Fixture: "kafka/rp-kafka-acls-list-yaml.golden"},
		{Args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBED --topic 'test-topic'", Fixture: "kafka/kafka-acls-invalid-operation.golden", WantErrCode: 1},
		{Args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{Args: "kafka acl create --cluster lkc-acls --allow --service-account sa-55555 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{Args: "kafka acl delete --cluster lkc-acls --allow --service-account sa-55555 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{Args: "kafka acl delete --cluster lkc-acls --allow --service-account sa-55555 --operation READ --operation DESCRIBE --topic 'test-topic'"},

		{Args: "kafka topic list --cluster lkc-kafka-api-topics", Login: "default", Fixture: "kafka/topic-list.golden"},
		//	{Args: "kafka topic list --cluster lkc-topics", Fixture: "kafka/rp-topic-list.golden"},
		{Args: "kafka topic list --cluster lkc-topics", Fixture: "kafka/topic-list.golden", Env: []string{"XX_CCLOUD_USE_KAFKA_API=true"}},
		{Args: "kafka topic list", Login: "default", UseKafka: "lkc-kafka-api-no-topics", Fixture: "kafka/topic-list-empty.golden"},
		{Args: "kafka topic list", Login: "default", UseKafka: "lkc-not-ready", Fixture: "kafka/cluster-not-ready.golden", WantErrCode: 1},

		{Args: "kafka topic create", Login: "default", UseKafka: "lkc-create-topic", Fixture: "kafka/topic-create.golden", WantErrCode: 1},
		{Args: "kafka topic create topic1", Login: "default", UseKafka: "lkc-create-topic-kafka-api", Fixture: "kafka/topic-create-success.golden"},
		{Args: "kafka topic create topic1", UseKafka: "lkc-create-topic", Fixture: "kafka/topic-create-success.golden"},
		//	{Args: "kafka topic create topic-exist", Login: "default", UseKafka: "lkc-create-topic", Fixture: "kafka/topic-create-dup-topic.golden", WantErrCode: 1},

		{Args: "kafka topic describe", Login: "default", UseKafka: "lkc-describe-topic", Fixture: "kafka/topic-describe.golden", WantErrCode: 1},
		{Args: "kafka topic describe topic1", Login: "default", UseKafka: "lkc-describe-topic-kafka-api", Fixture: "kafka/topic-describe-success.golden"},
		//	{Args: "kafka topic describe topic-exist", UseKafka: "lkc-describe-topic", Fixture: "kafka/rp-topic-describe-success.golden"},
		//	{Args: "kafka topic describe topic-exist --output json", Login: "default", UseKafka: "lkc-describe-topic", Fixture: "kafka/topic-describe-json-success.golden"},
		{Args: "kafka topic describe topic1 --cluster lkc-create-topic-kafka-api", Login: "default", Fixture: "kafka/topic-describe-not-found.golden", WantErrCode: 1},
		//	{Args: "kafka topic describe topic2", Login: "default", UseKafka: "lkc-describe-topic", Fixture: "kafka/topic2-describe-not-found.golden", WantErrCode: 1},

		{Args: "kafka topic delete", Login: "default", UseKafka: "lkc-delete-topic", Fixture: "kafka/topic-delete.golden", WantErrCode: 1},
		//	{Args: "kafka topic delete topic-exist", Login: "default", UseKafka: "lkc-delete-topic", Fixture: "kafka/topic-delete-success.golden"},
		//	{Args: "kafka topic delete topic-exist", UseKafka: "lkc-delete-topic", Fixture: "kafka/topic-delete-success.golden"},
		{Args: "kafka topic delete topic1 --cluster lkc-create-topic-kafka-api", Login: "default", Fixture: "kafka/topic-delete-not-found.golden", WantErrCode: 1},
		//	{Args: "kafka topic delete topic2", Login: "default", UseKafka: "lkc-delete-topic", Fixture: "kafka/topic2-delete-not-found.golden", WantErrCode: 1},

		{Args: "kafka topic update topic-exist --config retention.ms=1,compression.type=gzip", Login: "default", UseKafka: "lkc-describe-topic-kafka-api", Fixture: "kafka/topic-update-success.golden"},
		{Args: "kafka topic update topic-exist --config retention.ms=1,compression.type=gzip", UseKafka: "lkc-describe-topic", Fixture: "kafka/topic-update-success.golden"},

		// Cluster linking
		//{Args: "kafka link list --cluster lkc-describe-topic", Fixture: "kafka/cluster-linking/list-link-plain.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka link list --cluster lkc-describe-topic -o json", Fixture: "kafka/cluster-linking/list-link-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka link list --cluster lkc-describe-topic -o yaml", Fixture: "kafka/cluster-linking/list-link-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka link describe --cluster lkc-describe-topic link-1", Fixture: "kafka/cluster-linking/describe-link-plain.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka link describe --cluster lkc-describe-topic link-1 -o json", Fixture: "kafka/cluster-linking/describe-link-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka link describe --cluster lkc-describe-topic link-1 -o yaml", Fixture: "kafka/cluster-linking/describe-link-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},

		//{Args: "kafka mirror list --cluster lkc-describe-topic --link-Name link-1", Fixture: "kafka/cluster-linking/list-mirror.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror list --cluster lkc-describe-topic --link-Name link-1 -o json", Fixture: "kafka/cluster-linking/list-mirror-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror list --cluster lkc-describe-topic --link-Name link-1 -o yaml", Fixture: "kafka/cluster-linking/list-mirror-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror list --cluster lkc-describe-topic", Fixture: "kafka/cluster-linking/list-all-mirror.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror list --cluster lkc-describe-topic -o json", Fixture: "kafka/cluster-linking/list-all-mirror-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror list --cluster lkc-describe-topic -o yaml", Fixture: "kafka/cluster-linking/list-all-mirror-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror describe topic-1 --link-Name link-1 --cluster lkc-describe-topic", Fixture: "kafka/cluster-linking/describe-mirror.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror describe topic-1 --link-Name link-1 --cluster lkc-describe-topic -o json", Fixture: "kafka/cluster-linking/describe-mirror-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror describe topic-1 --link-Name link-1 --cluster lkc-describe-topic -o yaml", Fixture: "kafka/cluster-linking/describe-mirror-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link-Name link-1", Fixture: "kafka/cluster-linking/promote-mirror.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link-Name link-1 -o json", Fixture: "kafka/cluster-linking/promote-mirror-json.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		//{Args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link-Name link-1 -o yaml", Fixture: "kafka/cluster-linking/promote-mirror-yaml.golden", WantErrCode: 0, UseKafka: "lkc-describe-topic", Env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
	}

	ResetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.Login = "default"
		tt.Workflow = true
		s.RunCcloudTest(tt)
	}
}

//func (s *CLITestSuite) TestCCloudKafkaConsumerGroup() {
//	tests := []CLITest{
//		{Args: "kafka consumer-group --help", Fixture: "kafka/consumer-group-help.golden"},
//
//		{Args: "kafka consumer-group list --help", Fixture: "kafka/consumer-group-list.golden"},
//		{Args: "kafka consumer-group list", Fixture: "kafka/consumer-group-list-no-flag.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group list --cluster lkc-groups", Fixture: "kafka/consumer-group-list-success.golden", WantErrCode: 0},
//		{Args: "kafka consumer-group list --cluster lkc-groups -o json", Fixture: "kafka/consumer-group-list-success-json.golden"},
//		{Args: "kafka consumer-group list --cluster lkc-groups -o yaml", Fixture: "kafka/consumer-group-list-success-yaml.golden"},
//
//		{Args: "kafka consumer-group describe --help", Fixture: "kafka/consumer-group-describe.golden"},
//		{Args: "kafka consumer-group describe", Fixture: "kafka/consumer-group-describe-no-Args.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group describe --cluster lkc-groups", Fixture: "kafka/consumer-group-describe-no-Args.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group describe consumer-group-1", Fixture: "kafka/consumer-group-describe-no-flag.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group describe consumer-group-1 --cluster lkc-groups", Fixture: "kafka/consumer-group-describe-success.golden"},
//		{Args: "kafka consumer-group describe consumer-group-1 --cluster lkc-groups -o json", Fixture: "kafka/consumer-group-describe-success-json.golden"},
//		{Args: "kafka consumer-group describe consumer-group-1 --cluster lkc-groups -o yaml", Fixture: "kafka/consumer-group-describe-success-yaml.golden"},
//
//		{Args: "kafka consumer-group lag --help", Fixture: "kafka/consumer-group-lag-help.golden"},
//
//		{Args: "kafka consumer-group lag summarize --help", Fixture: "kafka/consumer-group-lag-summarize.golden"},
//		{Args: "kafka consumer-group lag summarize", Fixture: "kafka/consumer-group-lag-summarize-no-Args.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group lag summarize consumer-group-1 --cluster lkc-groups", Fixture: "kafka/consumer-group-lag-summarize-success.golden", WantErrCode: 0},
//		{Args: "kafka consumer-group lag summarize consumer-group-1 --cluster lkc-groups -o json", Fixture: "kafka/consumer-group-lag-summarize-success-json.golden"},
//		{Args: "kafka consumer-group lag summarize consumer-group-1 --cluster lkc-groups -o yaml", Fixture: "kafka/consumer-group-lag-summarize-success-yaml.golden"},
//
//		{Args: "kafka consumer-group lag list --help", Fixture: "kafka/consumer-group-lag-list.golden"},
//		{Args: "kafka consumer-group lag list", Fixture: "kafka/consumer-group-lag-list-no-Args.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group lag list consumer-group-1 --cluster lkc-groups", Fixture: "kafka/consumer-group-lag-list-success.golden"},
//		{Args: "kafka consumer-group lag list consumer-group-1 --cluster lkc-groups -o json", Fixture: "kafka/consumer-group-lag-list-success-json.golden"},
//		{Args: "kafka consumer-group lag list consumer-group-1 --cluster lkc-groups -o yaml", Fixture: "kafka/consumer-group-lag-list-success-yaml.golden"},
//
//		{Args: "kafka consumer-group lag get --help", Fixture: "kafka/consumer-group-lag-get.golden"},
//		{Args: "kafka consumer-group lag get", Fixture: "kafka/consumer-group-lag-get-no-Args.golden", WantErrCode: 1},
//		{Args: "kafka consumer-group lag get consumer-group-1 --cluster lkc-groups --topic topic-1 --partition 1", Fixture: "kafka/consumer-group-lag-get-success.golden"},
//		{Args: "kafka consumer-group lag get consumer-group-1 --cluster lkc-groups --topic topic-1 --partition 1 -o json", Fixture: "kafka/consumer-group-lag-get-success-json.golden"},
//		{Args: "kafka consumer-group lag get consumer-group-1 --cluster lkc-groups --topic topic-1 --partition 1 -o yaml", Fixture: "kafka/consumer-group-lag-get-success-yaml.golden"},
//	}
//	resetConfiguration(s.T(), "ccloud")
//	for _, tt := range tests {
//		tt.Login = "default"
//		tt.workflow = true
//		s.runCcloudTest(tt)
//	}
//}

func (s *CLITestSuite) TestConfluentKafkaTopicList() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Test correct usage
		{Args: fmt.Sprintf("kafka topic list --url %s --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/list.golden"},
		// Test with basic auth input
		{Args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("Miles\nTod\n"))}, Fixture: "kafka/confluent/topic/list-with-auth.golden"},
		{Args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), Login: "default", Fixture: "kafka/confluent/topic/list-with-auth-from-Login.golden"},
		{Args: fmt.Sprintf("kafka topic list --url %s --prompt", kafkaRestURL), Login: "default", PreCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("Miles\nTod\n"))}, Fixture: "kafka/confluent/topic/list-with-auth-prompt.golden"},
		// Test with CONFLUENT_REST_URL Env var
		{Args: "kafka topic list --no-auth", Fixture: "kafka/confluent/topic/list.golden", Env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Test failure when only one of client-cert-path or client-key-path are provided
		{Args: "kafka topic list --client-cert-path cert.crt", WantErrCode: 1, Fixture: "kafka/confluent/topic/client-cert-flag-error.golden", Env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		{Args: "kafka topic list --client-key-path cert.key", WantErrCode: 1, Fixture: "kafka/confluent/topic/client-cert-flag-error.golden", Env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Output should format correctly depending on format argument.
		{Args: fmt.Sprintf("kafka topic list --url %s -o human --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/list.golden"},
		{Args: fmt.Sprintf("kafka topic list --url %s -o yaml --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/list-yaml.golden"},
		{Args: fmt.Sprintf("kafka topic list --url %s -o json --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/list-json.golden"},
		// Invalid format string should throw error
		{Args: fmt.Sprintf("kafka topic list --url %s -o hello --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/list-output-error.golden", WantErrCode: 1, Name: "invalid format string should throw error"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicCreate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// <topic> errors
		{Args: fmt.Sprintf("kafka topic create --url %s --no-auth", kafkaRestURL), Contains: "Error: accepts 1 arg(s), received 0", WantErrCode: 1, Name: "missing topic-Name should return error"},
		{Args: fmt.Sprintf("kafka topic create topic-exist --url %s --no-auth", kafkaRestURL), Contains: "Error: topic \"topic-exist\" already exists for the Kafka cluster\n\nSuggestions:\n    To list topics for the cluster, use `confluent kafka topic list --url <url>`.", WantErrCode: 1, Name: "creating topic with existing topic Name should fail"},
		// --partitions errors
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions -2 --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Number of partitions must be larger than 0. (40002)\n", WantErrCode: 1, Name: "creating topic with negative partitions Name should fail"},
		// --replication-factor errors
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor 4 --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Replication factor: 4 larger than available brokers: 3. (40002)\n", WantErrCode: 1, Name: "creating topic with larger replication factor than num. brokers should fail"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor -2 --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Replication factor must be larger than 0. (40002)\n", WantErrCode: 1, Name: "creating topic with negative replication factor should fail"},
		// --config errors
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "creating topic with poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=1, --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "creating topic with poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=1,compression --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "creating topic with poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --config asdf=1 --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Unknown topic config Name: asdf (40002)\n", WantErrCode: 1, Name: "creating topic with incorrect config Name should fail"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=as --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Invalid value as for configuration retention.ms: Not a number of type LONG (40002)\n", WantErrCode: 1, Name: "creating topic with correct key incorrect config value should fail"},
		// Success
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/create-topic-success.golden", WantErrCode: 0, Name: "correct URL with default params (part 6, repl 3, no configs) should create successfully"},
		{Args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions 7 --replication-factor 2 --config retention.ms=100000,compression.type=gzip --no-auth", kafkaRestURL),
			Fixture: "kafka/confluent/topic/create-topic-success.golden", WantErrCode: 0, Name: "correct URL with valid optional params should create successfully"},
		// --ifnotexists
		{Args: fmt.Sprintf("kafka topic create topic-exist --url %s --if-not-exists --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/create-duplicate-topic-ifnotexists-success.golden", WantErrCode: 0, Name: "create topic with existing topic Name with if-not-exists flag should succeed"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicDelete() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{Args: fmt.Sprintf("kafka topic delete --url %s --no-auth", kafkaRestURL), Contains: "Error: accepts 1 arg(s), received 0", WantErrCode: 1, Name: "missing topic-Name should return error"},
		{Args: fmt.Sprintf("kafka topic delete topic-exist --url %s --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/delete-topic-success.golden", WantErrCode: 0, Name: "deleting existing topic with correct url should delete successfully"},
		{Args: fmt.Sprintf("kafka topic delete topic-not-exist --url %s --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/delete-topic-not-exist-failure.golden", WantErrCode: 1, Name: "deleting a non-existent topic should fail"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicUpdate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic Name errors
		{Args: fmt.Sprintf("kafka topic update --url %s --no-auth", kafkaRestURL), Contains: "Error: accepts 1 arg(s), received 0", WantErrCode: 1, Name: "missing topic-Name should return error"},
		{Args: fmt.Sprintf("kafka topic update topic-not-exist --url %s --no-auth", kafkaRestURL), Contains: "Error: REST request failed: This server does not host this topic-partition. (40403)\n", WantErrCode: 1, Name: "update config of a non-existent topic should fail"},
		// --config errors
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1, --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression --no-auth", kafkaRestURL), Contains: "Error: configuration must be in the form of key=value", WantErrCode: 1, Name: "poorly formatted config arg should fail"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config asdf=1 --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Config asdf cannot be found for TOPIC topic-exist in cluster cluster-1. (404)\n", WantErrCode: 1, Name: "incorrect config Name should fail"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=as --no-auth", kafkaRestURL), Contains: "Error: REST request failed: Invalid config value for resource ConfigResource(type=TOPIC, Name='topic-exist'): Invalid value as for configuration retention.ms: Not a number of type LONG (40002)\n", WantErrCode: 1, Name: "correct key incorrect config value should fail"},
		// Success cases
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/update-topic-config-success", WantErrCode: 0, Name: "valid config updates should succeed with configs printed sorted"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1000,retention.ms=1 --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/update-topic-config-duplicate-success", WantErrCode: 0, Name: "valid duplicate config should succeed with the later config value kept"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth -o json", kafkaRestURL), Fixture: "kafka/confluent/topic/update-topic-config-success-json.golden", WantErrCode: 0, Name: "config updates with json output"},
		{Args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-auth -o yaml", kafkaRestURL), Fixture: "kafka/confluent/topic/update-topic-config-success-yaml.golden", WantErrCode: 0, Name: "config updates with yaml output"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaTopicDescribe() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic Name errors
		{Args: fmt.Sprintf("kafka topic describe --url %s --no-auth", kafkaRestURL), Contains: "Error: accepts 1 arg(s), received 0", WantErrCode: 1, Name: "<topic> arg missing should lead to error"},
		{Args: fmt.Sprintf("kafka topic describe topic-not-exist --url %s --no-auth", kafkaRestURL), Contains: "Error: REST request failed: This server does not host this topic-partition. (40403)\n", WantErrCode: 1, Name: "describing a non-existant topic should lead to error"},
		// -o errors
		{Args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o asdf --no-auth", kafkaRestURL), Contains: "Error: invalid value \"asdf\" for flag `--output`\n\nSuggestions:\n    The possible values for flag `output` are: human, json, yaml.", WantErrCode: 1, Name: "bad output format flag should lead to error"},
		// Success cases
		{Args: fmt.Sprintf("kafka topic describe topic-exist --url %s --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/describe-topic-success.golden", WantErrCode: 0, Name: "topic that exists & correct format arg should lead to success"},
		{Args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o human --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/describe-topic-success.golden", WantErrCode: 0, Name: "topic that exist & human arg should lead to success"},
		{Args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o json --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/describe-topic-success-json.golden", WantErrCode: 0, Name: "topic that exist & json arg should lead to success"},
		{Args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o yaml --no-auth", kafkaRestURL), Fixture: "kafka/confluent/topic/describe-topic-success-yaml.golden", WantErrCode: 0, Name: "topic that exist & yaml arg should lead to success"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}

func (s *CLITestSuite) TestConfluentKafkaACL() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// error case: bad operation, specified more than one resource type
		{Args: fmt.Sprintf("kafka acl list --operation fake --topic Test --consumer-group Group:Test --url %s --no-auth", kafkaRestURL), Name: "bad operation and conflicting resource type errors", Fixture: "kafka/confluent/acl/list-errors.golden", WantErrCode: 1},
		// success cases
		{Args: fmt.Sprintf("kafka acl list --url %s --no-auth", kafkaRestURL), Name: "acl list output human", Fixture: "kafka/confluent/acl/acl-list.golden"},
		{Args: fmt.Sprintf("kafka acl list -o json --url %s --no-auth", kafkaRestURL), Name: "acl list output json", Fixture: "kafka/confluent/acl/acl-list-json.golden"},
		{Args: fmt.Sprintf("kafka acl list -o yaml --url %s --no-auth", kafkaRestURL), Name: "acl list output yaml", Fixture: "kafka/confluent/acl/acl-list-yaml.golden"},

		// error case: bad operation, specified more than one resource type, allow/deny not set
		{Args: fmt.Sprintf("kafka acl create --principal User:Alice --operation fake --topic Test --consumer-group Group:Test --url %s --no-auth", kafkaRestURL), Name: "bad operation, conflicting resource type, no allow/deny specified errors", Fixture: "kafka/confluent/acl/create-errors.golden", WantErrCode: 1},
		// success cases
		{Args: fmt.Sprintf("kafka acl create --operation write --cluster-scope --principal User:Alice --allow --url %s --no-auth", kafkaRestURL), Name: "acl create output human", Fixture: "kafka/confluent/acl/acl-create.golden"},
		{Args: fmt.Sprintf("kafka acl create --operation all --cluster-scope --principal User:Alice --allow -o json --url %s --no-auth", kafkaRestURL), Name: "acl create output json", Fixture: "kafka/confluent/acl/acl-create-json.golden"},
		{Args: fmt.Sprintf("kafka acl create --operation all --topic Test --principal User:Alice --allow -o yaml --url %s --no-auth", kafkaRestURL), Name: "acl create output yaml", Fixture: "kafka/confluent/acl/acl-create-yaml.golden"},

		// error case: bad operation, specified more than one resource type, allow/deny not set
		{Args: fmt.Sprintf("kafka acl delete --principal User:Alice --host '*' --operation fake --topic Test --consumer-group Group:Test --url %s --no-auth", kafkaRestURL), Name: "bad operation, conflicting resource type, no allow/deny specified errors", Fixture: "kafka/confluent/acl/delete-errors.golden", WantErrCode: 1},
		// success cases
		{Args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation READ --principal User:Alice --allow --url %s --no-auth", kafkaRestURL), Name: "acl delete output human", Fixture: "kafka/confluent/acl/acl-delete.golden"},
		{Args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation READ --principal User:Alice --allow -o json --url %s --no-auth", kafkaRestURL), Name: "acl delete output json", Fixture: "kafka/confluent/acl/acl-delete-json.golden"},
		{Args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation READ --principal User:Alice --allow -o yaml --url %s --no-auth", kafkaRestURL), Name: "acl delete output yaml", Fixture: "kafka/confluent/acl/acl-delete-yaml.golden"},
	}

	for _, clitest := range tests {
		s.RunConfluentTest(clitest)
	}
}
