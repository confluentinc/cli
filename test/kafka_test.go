package test

import (
	"fmt"
	"os"
	"runtime"
)

func (s *CLITestSuite) TestKafka() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	createLinkConfigFile := getCreateLinkConfigFile()
	defer os.Remove(createLinkConfigFile)
	createBidirectionalLinkConfigFile := getCreateBidirectionalLinkConfigFile()
	defer os.Remove(createBidirectionalLinkConfigFile)
	tests := []CLITest{
		{args: "environment use env-596", fixture: "kafka/0.golden"},
		{args: "kafka cluster list", fixture: "kafka/6.golden"},
		{args: "kafka cluster list -o json", fixture: "kafka/7.golden"},
		{args: "kafka cluster list -o yaml", fixture: "kafka/8.golden"},

		{args: "environment use env-123", fixture: "kafka/46.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone", fixture: "kafka/2.golden"},
		{args: "kafka cluster list", fixture: "kafka/6.golden"},
		{args: "kafka cluster list --all", fixture: "kafka/47.golden"},

		{args: "environment use env-596", fixture: "kafka/0.golden"},
		{args: "kafka cluster create", fixture: "kafka/1.golden", exitCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone", fixture: "kafka/2.golden"},
		{args: "kafka cluster create my-failed-cluster --cloud oops --region us-east1 --availability single-zone", fixture: "kafka/cluster/create-cloud-provider-error.golden", exitCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region oops --availability single-zone", fixture: "kafka/cluster/create-cloud-region-error.golden", exitCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type oops", fixture: "kafka/cluster/create-type-error.golden", exitCode: 1},
		{args: "kafka cluster create my-failed-cluster --cloud aws --region us-east-1 --availability single-zone --type dedicated --cku 0", fixture: "kafka/cluster/create-cku-error.golden", exitCode: 1},
		{args: "kafka cluster create my-dedicated-cluster --cloud aws --region us-east-1 --type dedicated --cku 1", fixture: "kafka/22.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o json", fixture: "kafka/23.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability single-zone -o yaml", fixture: "kafka/24.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --availability oops-zone", fixture: "kafka/cluster/create-availability-zone-error.golden", exitCode: 1},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --type enterprise --availability multi-zone", fixture: "kafka/cluster/create-enterprise.golden"},
		{args: "kafka cluster create my-new-cluster --cloud aws --region us-east-1 --type enterprise", fixture: "kafka/cluster/create-enterprise-availability-zone-error.golden", exitCode: 1},

		{args: "kafka cluster update lkc-update", fixture: "kafka/cluster/create-flag-error.golden", exitCode: 1},
		{args: "kafka cluster update lkc-update --name lkc-update-name", fixture: "kafka/26.golden"},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o json", fixture: "kafka/28.golden"},
		{args: "kafka cluster update lkc-update --name lkc-update-name -o yaml", fixture: "kafka/29.golden"},
		{args: "kafka cluster update lkc-update-dedicated-expand --name lkc-update-dedicated-name --cku 2", fixture: "kafka/27.golden"},
		{args: "kafka cluster update lkc-update-dedicated-expand --cku 2", fixture: "kafka/39.golden"},
		{args: "kafka cluster update lkc-update --cku 2", fixture: "kafka/cluster/update-resize-error.golden", exitCode: 1},
		{args: "kafka cluster update lkc-update-dedicated-shrink --name lkc-update-dedicated-name --cku 1", fixture: "kafka/44.golden"},
		{args: "kafka cluster update lkc-update-dedicated-shrink --cku 1", fixture: "kafka/45.golden"},
		{args: "kafka cluster update lkc-update-dedicated-shrink-multi --cku 1", fixture: "kafka/cluster/update-dedicated-shrink-error.golden", exitCode: 1},
		{args: "kafka cluster update lkc-update --cku 1", fixture: "kafka/cluster/update-resize-error.golden", exitCode: 1},

		{args: "kafka cluster delete --force", fixture: "kafka/3.golden", exitCode: 1},
		{args: "kafka cluster delete lkc-unknown --force", fixture: "kafka/cluster/delete-unknown-error.golden", exitCode: 1},
		{args: "kafka cluster delete lkc-def973 --force", fixture: "kafka/5.golden"},
		{args: "kafka cluster delete lkc-def973", input: "kafka-cluster\n", fixture: "kafka/5-prompt.golden"},
		{args: "kafka cluster delete lkc-def973 lkc-unknown", fixture: "kafka/cluster/delete-multiple-fail.golden", exitCode: 1},
		{args: "kafka cluster delete lkc-def973 lkc-update", input: "n\n", fixture: "kafka/cluster/delete-multiple-refuse.golden"},
		{args: "kafka cluster delete lkc-def973 lkc-update", input: "y\n", fixture: "kafka/cluster/delete-multiple-success.golden"},

		{args: "kafka cluster use lkc-12345", fixture: "kafka/40.golden"},

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

		{args: "kafka cluster describe lkc-describe-dedicated-provisioning", fixture: "kafka/cluster-describe-dedicated-provisioning.golden"},

		{args: "kafka cluster describe lkc-unknown", fixture: "kafka/48.golden", exitCode: 1},
		{args: "kafka cluster describe lkc-unknown-type", fixture: "kafka/describe-unknown-cluster-type.golden"},

		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka/acl/list-cloud.golden"},
		{args: "kafka acl list --cluster lkc-acls --all", fixture: "kafka/acl/list-cloud-all.golden"},
		{args: "kafka acl list --cluster lkc-acls -o json", fixture: "kafka/acl/list-json-cloud.golden"},
		{args: "kafka acl list --cluster lkc-acls -o yaml", fixture: "kafka/acl/list-yaml-cloud.golden"},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operations read,described --topic test-topic", fixture: "kafka/acl/invalid-operation.golden", exitCode: 1},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account sa-12345 --operations read,describe --topic test-topic", fixture: "kafka/acl/create-service-account.golden"},
		{args: "kafka acl create --cluster lkc-acls --allow --principal User:sa-12345 --operations write,alter --topic test-topic", fixture: "kafka/acl/create-principal.golden"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account sa-12345 --operations read,describe --topic test-topic --force", fixture: "kafka/acl/delete-cloud.golden"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account sa-12345 --operations read,describe --topic test-topic", input: "y\n", fixture: "kafka/acl/delete-cloud-prompt.golden"},
		{args: "kafka acl delete --cluster lkc-acls --allow --principal User:sa-12345 --operations write,alter --topic test-topic --force", fixture: "kafka/acl/delete-cloud.golden"},

		{args: "kafka topic list --cluster lkc-kafka-api-topics", login: "cloud", fixture: "kafka/topic/list-cloud.golden"},
		{args: "kafka topic list --cluster lkc-topics", fixture: "kafka/topic/list-cloud.golden"},

		{args: "kafka topic create", login: "cloud", useKafka: "lkc-create-topic", fixture: "kafka/topic/create.golden", exitCode: 1},
		{args: "kafka topic create topic1", useKafka: "lkc-create-topic", fixture: "kafka/topic/create-success.golden"},
		{args: "kafka topic create topic1 --dry-run", useKafka: "lkc-create-topic", fixture: "kafka/topic/create-success.golden"},
		{args: "kafka topic create topic-exist", login: "cloud", useKafka: "lkc-create-topic", fixture: "kafka/topic/create-dup-topic.golden", exitCode: 1},
		{args: "kafka topic create topic-exceed-limit --partitions 9001", login: "cloud", useKafka: "lkc-create-topic", fixture: "kafka/topic/create-limit-topic.golden", exitCode: 1},

		{args: "kafka topic describe", login: "cloud", useKafka: "lkc-describe-topic", fixture: "kafka/topic/describe.golden", exitCode: 1},
		{args: "kafka topic describe topic-exist", useKafka: "lkc-describe-topic", fixture: "kafka/topic/describe-success.golden"},
		{args: "kafka topic describe topic-exist --output json", login: "cloud", useKafka: "lkc-describe-topic", fixture: "kafka/topic/describe-json-success.golden"},
		{args: "kafka topic describe topic2", login: "cloud", useKafka: "lkc-describe-topic", fixture: "kafka/topic/describe-not-found-topic2.golden", exitCode: 1},

		{args: "kafka topic delete --force", login: "cloud", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete.golden", exitCode: 1},
		{args: "kafka topic delete topic-exist --force", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete-success.golden"},
		{args: "kafka topic delete topic-exist", useKafka: "lkc-delete-topic", input: "topic-exist\n", fixture: "kafka/topic/delete-success-prompt.golden"},
		{args: "kafka topic delete topic-exist topic2", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete-multiple-fail.golden", exitCode: 1},
		{args: "kafka topic delete topic-exist topic-exist-2", input: "n\n", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete-multiple-refuse.golden"},
		{args: "kafka topic delete topic-exist topic-exist-2", input: "y\n", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete-multiple-success.golden"},
		{args: "kafka topic delete topic2 --force", login: "cloud", useKafka: "lkc-delete-topic", fixture: "kafka/topic/delete-not-found-topic2.golden", exitCode: 1},

		{args: "kafka topic update topic-exist-rest --config retention.ms=1,compression.type=gzip", useKafka: "lkc-describe-topic", fixture: "kafka/topic/update-success-rest.golden"},
		{args: "kafka topic update topic-exist-rest --config retention.ms=1,compression.type=gzip --dry-run", useKafka: "lkc-describe-topic", fixture: "kafka/topic/update-success-dry-run.golden"},
		{args: "kafka topic update topic-exist-rest --config retention.ms=1,compression.type=gzip -o json", useKafka: "lkc-describe-topic", fixture: "kafka/topic/update-success-rest-json.golden"},
		{args: "kafka topic update topic-exist-rest --config retention.ms=1,compression.type=gzip -o yaml", useKafka: "lkc-describe-topic", fixture: "kafka/topic/update-success-rest-yaml.golden"},
		{args: "kafka topic update topic-exist-rest --config num.partitions=6", useKafka: "lkc-describe-topic", fixture: "kafka/topic/update-success-rest-partitions-count.golden"},
	}

	if runtime.GOOS != "windows" {
		noSchemaTest := CLITest{args: "kafka topic produce topic-exist --value-format avro --api-key key --api-secret secret", login: "cloud", useKafka: "lkc-create-topic", fixture: "kafka/topic/produce-no-schema.golden", exitCode: 1}
		tests = append(tests, noSchemaTest)
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}

	tests = []CLITest{
		{args: fmt.Sprintf("kafka link describe link-1 --url %s", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/link/describe-onprem.golden"},
		{args: fmt.Sprintf("kafka link task list link-5 --url %s", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/link/list-tasks-onprem.golden"},
		{args: fmt.Sprintf("kafka link task list link-5 --url %s -o yaml", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/link/list-tasks-onprem-yaml.golden"},
		{args: fmt.Sprintf("kafka link task list link-5 --url %s -o json", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/link/list-tasks-onprem-json.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaClusterCreate_Byok() {
	test := CLITest{
		login:   "cloud",
		args:    "kafka cluster create cck-byok-test --cloud aws --region us-east-1 --type dedicated --cku 1 --byok cck-001",
		fixture: "kafka/cluster/cck-byok.golden",
	}

	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestKafkaClusterConfiguration() {
	tests := []CLITest{
		{args: "kafka cluster use lkc-12345"},
		{args: "kafka cluster configuration describe compression.type", fixture: "kafka/cluster/configuration/describe.golden"},
		{args: "kafka cluster configuration update --config auto.create.topics.enable=true", fixture: "kafka/cluster/configuration/update.golden"},
		{args: `kafka cluster configuration update --config ssl.cipher.suites="val1,val2,val3"`, fixture: "kafka/cluster/configuration/update-with-commas.golden"},
		{args: "kafka cluster configuration update --config test/fixtures/input/kafka/cluster/configuration/update.properties", fixture: "kafka/cluster/configuration/update.golden"},
		{args: "kafka cluster configuration list", fixture: "kafka/cluster/configuration/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaClientConfig() {
	// TODO: add --config flag to all commands or ENVVAR instead of using standard config file location
	tests := []CLITest{
		// error - missing context cluster
		{args: "kafka client-config create java", fixture: "kafka/client-config/no-cluster.golden", exitCode: 1},
		// error - missing context kafka key-secret pair
		{args: "kafka client-config create java", useKafka: "lkc-cool1", fixture: "kafka/client-config/no-keypair.golden", exitCode: 1},

		// set kafka key-secret pair
		{args: "api-key store UIAPIKEY100 UIAPISECRET100 --resource lkc-cool1"},
		{args: "api-key use UIAPIKEY100"},

		// warning - missing sr key-secret pair
		{args: "kafka client-config create java", useKafka: "lkc-cool1", fixture: "kafka/client-config/java-no-sr-keypair.golden"},

		// pass - does not need sr key-secret pair
		{args: "kafka client-config create csharp", useKafka: "lkc-cool1", fixture: "kafka/client-config/csharp.golden"},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.login = "cloud"
		test.workflow = true
		s.runIntegrationTest(test)
	}
}

func getCreateLinkConfigFile() string {
	file, _ := os.CreateTemp(os.TempDir(), "test")
	_, _ = file.Write([]byte("key=val\n key2=val2 \n key3=val password=pass"))
	return file.Name()
}

func getCreateBidirectionalLinkConfigFile() string {
	file, _ := os.CreateTemp(os.TempDir(), "test")
	_, _ = file.Write([]byte("link.mode=BIDIRECTIONAL \nkey=val\n key2=val2 \n key3=val password=pass"))
	return file.Name()
}

func (s *CLITestSuite) TestKafkaBroker() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: "kafka broker describe 1", fixture: "kafka/broker/describe-1.golden"},
		{args: "kafka broker describe 1 -o json", fixture: "kafka/broker/describe-1-json.golden"},
		{args: "kafka broker describe 1 -o yaml", fixture: "kafka/broker/describe-1-yaml.golden"},
		{args: "kafka broker describe --all", fixture: "kafka/broker/describe-all.golden"},
		{args: "kafka broker describe --all -o json", fixture: "kafka/broker/describe-all-json.golden"},
		{args: "kafka broker describe --all -o yaml", fixture: "kafka/broker/describe-all-yaml.golden"},
		{args: "kafka broker describe 1 --config-name compression.type", fixture: "kafka/broker/describe-1-config.golden"},
		{args: "kafka broker describe --all --config-name compression.type", fixture: "kafka/broker/describe-all-config.golden"},
		{args: "kafka broker describe --all --config-name compression.type -o json", fixture: "kafka/broker/describe-all-config-json.golden"},
		{args: "kafka broker describe 1 --all", exitCode: 1, fixture: "kafka/broker/err-all-and-arg.golden"},

		{args: "kafka broker update --all --config compression.type=zip,sasl_mechanism=SASL/PLAIN", fixture: "kafka/broker/update-all.golden"},
		{args: "kafka broker update 1 --config test/fixtures/input/kafka/broker/update.properties", fixture: "kafka/broker/update-1.golden"},
		{args: "kafka broker update --config compression.type=zip,sasl_mechanism=SASL/PLAIN", exitCode: 1, fixture: "kafka/broker/err-need-all-or-arg.golden"},

		{args: "kafka broker delete 1 --force", fixture: "kafka/broker/delete.golden"},
		{args: "kafka broker delete 1", input: "y\n", fixture: "kafka/broker/delete-prompt.golden"},
		{args: "kafka broker delete 1 2", input: "n\n", fixture: "kafka/broker/delete-multiple-refuse.golden"},
		{args: "kafka broker delete 1 2", input: "y\n", fixture: "kafka/broker/delete-multiple-prompt.golden"},
		{args: "kafka broker delete 1 3", fixture: "kafka/broker/delete-multiple-fail.golden", exitCode: 1},

		{args: "kafka broker get-tasks 1", fixture: "kafka/broker/get-tasks-1.golden"},
		{args: "kafka broker get-tasks 1 --task-type remove-broker", fixture: "kafka/broker/get-tasks-1-remove-broker.golden"},
		{args: "kafka broker get-tasks --all", fixture: "kafka/broker/get-tasks-all.golden"},
		{args: "kafka broker get-tasks --all --task-type add-broker", fixture: "kafka/broker/get-tasks-all-add-broker.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		test.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaBrokerList() {
	tests := []CLITest{
		{args: fmt.Sprintf("kafka broker list --url %s", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/broker/list.golden"},
		{args: fmt.Sprintf("kafka broker list --url %s", s.TestBackend.GetCloudUrl()), fixture: "kafka/broker/list-warning-cloud-url.golden", exitCode: 1},
		{args: fmt.Sprintf("kafka broker list --url %s -o json", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/broker/list-json.golden"},
		{args: fmt.Sprintf("kafka broker list --url %s -o yaml", s.TestBackend.GetKafkaRestUrl()), fixture: "kafka/broker/list-yaml.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaLink() {
	tests := []CLITest{
		{args: "kafka link configuration list --cluster lkc-describe-topic link-1 -o yaml", fixture: "kafka/link/configuration/list-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link configuration list --cluster lkc-describe-topic link-1", fixture: "kafka/link/configuration/list.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link configuration list --cluster lkc-describe-topic link-4 -o yaml", fixture: "kafka/link/configuration/list-bidirectional-link-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link configuration list --cluster lkc-describe-topic link-4", fixture: "kafka/link/configuration/list-bidirectional-link.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link create bidirectional_link --remote-cluster lkc-describe-topic --local-api-key local-api-key123 --local-api-secret local-api-secret-123 --remote-api-key remote-api-key-123 --remote-api-secret remote-api-secret-123 --remote-bootstrap-server myhost:1234 --config " + getCreateBidirectionalLinkConfigFile(), fixture: "kafka/link/create-bidirectional-link.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link create my_link --source-cluster lkc-describe-topic --source-bootstrap-server myhost:1234 --config " + getCreateLinkConfigFile(), fixture: "kafka/link/create-link.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link delete link-1 link-2 link-dne", fixture: "kafka/link/delete-link-multiple-fail.golden", useKafka: "lkc-describe-topic", exitCode: 1},
		{args: "kafka link delete link-1 link-2", input: "n\n", fixture: "kafka/link/delete-link-multiple-refuse.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link delete link-1 link-2", input: "y\n", fixture: "kafka/link/delete-link-multiple.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link delete link-1", input: "link-1\n", fixture: "kafka/link/delete-link.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link delete myLink_1", input: "myLink_1\n", fixture: "kafka/link/delete-link-no-prefix.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link describe link-1 --cluster lkc-describe-topic -o json", fixture: "kafka/link/describe-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link describe link-1 --cluster lkc-describe-topic", fixture: "kafka/link/describe.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link describe link-3 --cluster lkc-describe-topic", fixture: "kafka/link/describe-error.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link describe link-4 --cluster lkc-describe-topic", fixture: "kafka/link/describe-bidirectional-link.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link task list link-5 --cluster lkc-describe-topic", fixture: "kafka/link/list-link-with-tasks.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link task list link-5 --cluster lkc-describe-topic -o json", fixture: "kafka/link/list-link-with-tasks-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link task list link-5 --cluster lkc-describe-topic -o yaml", fixture: "kafka/link/list-link-with-tasks-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link list --cluster lkc-describe-topic -o json", fixture: "kafka/link/list-link-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link list --cluster lkc-describe-topic -o yaml", fixture: "kafka/link/list-link-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka link list --cluster lkc-describe-topic", fixture: "kafka/link/list-link-plain.golden", useKafka: "lkc-describe-topic"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaMirror() {
	tests := []CLITest{
		{args: "kafka mirror describe topic-1 --link link-1 --cluster lkc-describe-topic -o json", fixture: "kafka/mirror/describe-mirror-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror describe topic-1 --link link-1 --cluster lkc-describe-topic -o yaml", fixture: "kafka/mirror/describe-mirror-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror describe topic-1 --link link-1 --cluster lkc-describe-topic", fixture: "kafka/mirror/describe-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror state-transition-error list topic-1 --link link-1 --cluster lkc-describe-topic", fixture: "kafka/mirror/list-mirror-with-state-transition-errors.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror state-transition-error list topic-1 --link link-1 --cluster lkc-describe-topic -o json", fixture: "kafka/mirror/list-mirror-with-state-transition-errors-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror state-transition-error list topic-1 --link link-1 --cluster lkc-describe-topic -o yaml", fixture: "kafka/mirror/list-mirror-with-state-transition-errors-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror failover topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/failover-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic --link link-1 -o json", fixture: "kafka/mirror/list-mirror-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic --link link-1 -o yaml", fixture: "kafka/mirror/list-mirror-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/list-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic -o json", fixture: "kafka/mirror/list-all-mirror-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic -o yaml", fixture: "kafka/mirror/list-all-mirror-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror list --cluster lkc-describe-topic", fixture: "kafka/mirror/list-all-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror pause topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/pause-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link link-1 -o json", fixture: "kafka/mirror/promote-mirror-json.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link link-1 -o yaml", fixture: "kafka/mirror/promote-mirror-yaml.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror promote topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/promote-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror resume topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/resume-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror reverse-and-start topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/reverse-and-start-mirror.golden", useKafka: "lkc-describe-topic"},
		{args: "kafka mirror reverse-and-pause topic1 topic2 --cluster lkc-describe-topic --link link-1", fixture: "kafka/mirror/reverse-and-pause-mirror.golden", useKafka: "lkc-describe-topic"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaPartition() {
	tests := []CLITest{
		{args: "kafka partition describe 0 --topic topic1 --cluster lkc-12345", fixture: "kafka/partition/describe.golden"},
		{args: "kafka partition describe 0 --topic topic1 --cluster lkc-12345 -o json", fixture: "kafka/partition/describe-json.golden"},
		{args: "kafka partition list --topic topic1 --cluster lkc-12345", fixture: "kafka/partition/list.golden"},
		{args: "kafka partition list --topic topic1 --cluster lkc-12345 -o json", fixture: "kafka/partition/list-json.golden"},
	}
	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests = []CLITest{
		{args: "kafka partition list --topic topic1", fixture: "kafka/partition/list-onprem.golden"},
		{args: "kafka partition list --topic topic1 -o json", fixture: "kafka/partition/list-json-onprem.golden"},
		{args: "kafka partition list --topic topic1 -o yaml", fixture: "kafka/partition/list-yaml-onprem.golden"},
		{args: "kafka partition describe 0 --topic topic1", fixture: "kafka/partition/describe-onprem.golden"},
		{args: "kafka partition describe 0 --topic topic1 -o json", fixture: "kafka/partition/describe-json-onprem.golden"},
		{args: "kafka partition describe 0 --topic topic1 -o yaml", fixture: "kafka/partition/describe-yaml-onprem.golden"},
		{args: "kafka partition reassignment list", fixture: "kafka/partition/reassignment/list.golden"},
		{args: "kafka partition reassignment list -o json", fixture: "kafka/partition/reassignment/list-json.golden"},
		{args: "kafka partition reassignment list --topic topic1", fixture: "kafka/partition/reassignment/list-by-topic.golden"},
		{args: "kafka partition reassignment list 0 --topic topic1", fixture: "kafka/partition/reassignment/list-by-partition.golden"},
		{args: "kafka partition reassignment list 0 --topic topic1 -o yaml", fixture: "kafka/partition/reassignment/list-by-partition-yaml.golden"},
	}
	for _, test := range tests {
		test.login = "onprem"
		test.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaReplica() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: "kafka replica list --topic topic-exist", fixture: "kafka/replica/list-topic-replicas.golden"},
		{args: "kafka replica list --topic topic-exist -o json", fixture: "kafka/replica/list-topic-replicas-json.golden"},
		{args: "kafka replica list --topic topic-exist --partition 2", fixture: "kafka/replica/list-partition-replicas.golden"},
		{args: "kafka replica list --topic topic-exist --partition 2 -o yaml", fixture: "kafka/replica/list-partition-replicas-yaml.golden"},
		{args: "kafka replica list", fixture: "kafka/replica/no-flags-error.golden", exitCode: 1},
	}
	for _, test := range tests {
		test.login = "onprem"
		test.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaTopicList() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Test correct usage
		{args: fmt.Sprintf("kafka topic list --url %s --no-authentication", kafkaRestURL), fixture: "kafka/topic/list.golden"},
		// Test with basic auth input
		{args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), input: "Miles\nTod\n", fixture: "kafka/topic/list-with-auth.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s", kafkaRestURL), login: "onprem", fixture: "kafka/topic/list-with-auth-from-login.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s --prompt", kafkaRestURL), login: "onprem", input: "Miles\nTod\n", fixture: "kafka/topic/list-with-auth-prompt.golden"},
		// Test with CONFLUENT_REST_URL env var
		{args: "kafka topic list --no-authentication", fixture: "kafka/topic/list.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Test failure when only one of client-cert-path or client-key-path are provided
		{args: "kafka topic list --client-cert-path cert.crt", exitCode: 1, fixture: "kafka/topic/client-cert-flag-error.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		{args: "kafka topic list --client-key-path cert.key", exitCode: 1, fixture: "kafka/topic/client-cert-flag-error.golden", env: []string{"CONFLUENT_REST_URL=" + kafkaRestURL}},
		// Output should format correctly depending on format argument.
		{args: fmt.Sprintf("kafka topic list --url %s -o human --no-authentication", kafkaRestURL), fixture: "kafka/topic/list.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o yaml --no-authentication", kafkaRestURL), fixture: "kafka/topic/list-yaml.golden"},
		{args: fmt.Sprintf("kafka topic list --url %s -o json --no-authentication", kafkaRestURL), fixture: "kafka/topic/list-json.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaTopicCreate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// <topic> errors
		{args: fmt.Sprintf("kafka topic create --url %s --no-authentication", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", exitCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic create topic-exist --url %s --no-authentication", kafkaRestURL), contains: "Error: topic \"topic-exist\" already exists for the Kafka cluster\n\nSuggestions:\n    To list topics for the cluster, use `confluent kafka topic list --url <url>`.", exitCode: 1, name: "creating topic with existing topic name should fail"},
		// --partitions errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions -2 --no-authentication", kafkaRestURL), exitCode: 1, name: "creating topic with negative partitions name should fail", fixture: "kafka/topic/create-negative-partitions.golden"},
		// --replication-factor errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor 4 --no-authentication", kafkaRestURL), contains: "Error: REST request failed: Replication factor: 4 larger than available brokers: 3.\n", exitCode: 1, name: "creating topic with larger replication factor than num. brokers should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --replication-factor -2 --no-authentication", kafkaRestURL), exitCode: 1, name: "creating topic with negative replication factor should fail", fixture: "kafka/topic/create-negative-replication-factor.golden"},
		// --config errors
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config asdf=1 --no-authentication", kafkaRestURL), contains: "Error: REST request failed: Unknown topic config name: asdf\n", exitCode: 1, name: "creating topic with incorrect config name should fail"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --config retention.ms=as --no-authentication", kafkaRestURL), contains: "Error: REST request failed: Invalid value as for configuration retention.ms: Not a number of type LONG\n", exitCode: 1, name: "creating topic with correct key incorrect config value should fail"},
		// Success
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --no-authentication", kafkaRestURL), fixture: "kafka/topic/create-topic-success.golden", name: "correct URL with default params (part 6, repl 3, no configs) should create successfully"},
		{args: fmt.Sprintf("kafka topic create topic-X --url %s --partitions 7 --replication-factor 2 --config retention.ms=100000,compression.type=gzip --no-authentication", kafkaRestURL), fixture: "kafka/topic/create-topic-success.golden", name: "correct URL with valid optional params should create successfully"},
		// --ifnotexists
		{args: fmt.Sprintf("kafka topic create topic-exist --url %s --if-not-exists --no-authentication", kafkaRestURL), fixture: "kafka/topic/create-duplicate-topic-ifnotexists-success.golden", name: "create topic with existing topic name with if-not-exists flag should succeed"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaTopicDelete() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		{args: fmt.Sprintf("kafka topic delete --url %s --no-authentication --force", kafkaRestURL), contains: "Error: requires at least 1 arg(s), only received 0", exitCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic delete topic-exist --url %s --no-authentication --force", kafkaRestURL), fixture: "kafka/topic/delete-topic-success.golden", name: "deleting existing topic with correct url should delete successfully"},
		{args: fmt.Sprintf("kafka topic delete topic-exist --url %s --no-authentication", kafkaRestURL), input: "topic-exist\n", fixture: "kafka/topic/delete-topic-success-prompt.golden", name: "deleting existing topic with correct url and prompt should delete successfully"},
		{args: fmt.Sprintf("kafka topic delete topic-exist topic2 --url %s --no-authentication", kafkaRestURL), fixture: "kafka/topic/delete-multiple-fail.golden", exitCode: 1, name: "deleting existing topic and non-existent topic should fail"},
		{args: fmt.Sprintf("kafka topic delete topic-exist topic-exist-2 --url %s --no-authentication", kafkaRestURL), input: "n\n", fixture: "kafka/topic/delete-multiple-refuse.golden", name: "answering no should prevent topic deletion"},
		{args: fmt.Sprintf("kafka topic delete topic-exist topic-exist-2 --url %s --no-authentication", kafkaRestURL), input: "y\n", fixture: "kafka/topic/delete-multiple-success.golden", name: "deleting multiple existing topics should succeed"},
		{args: fmt.Sprintf("kafka topic delete topic-not-exist --url %s --no-authentication --force", kafkaRestURL), fixture: "kafka/topic/delete-topic-not-exist-failure.golden", exitCode: 1, name: "deleting a non-existent topic should fail"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaTopicUpdate() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic name errors
		{args: fmt.Sprintf("kafka topic update --url %s --no-authentication", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", exitCode: 1, name: "missing topic-name should return error"},
		{args: fmt.Sprintf("kafka topic update topic-not-exist --url %s --no-authentication", kafkaRestURL), contains: "Error: REST request failed: This server does not host this topic-partition.\n", exitCode: 1, name: "update config of a non-existent topic should fail"},
		// --config errors
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config asdf=1 --no-authentication", kafkaRestURL), contains: "Error: REST request failed: Config asdf cannot be found for TOPIC topic-exist in cluster cluster-1.\n", exitCode: 1, name: "incorrect config name should fail"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=as --no-authentication", kafkaRestURL), contains: "Error: REST request failed: Invalid config value for resource ConfigResource(type=TOPIC, name='topic-exist'): Invalid value as for configuration retention.ms: Not a number of type LONG\n", exitCode: 1, name: "correct key incorrect config value should fail"},
		// Success cases
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-authentication", kafkaRestURL), fixture: "kafka/topic/update-topic-config-success", name: "valid config updates should succeed with configs printed sorted"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1000,retention.ms=1 --no-authentication", kafkaRestURL), fixture: "kafka/topic/update-topic-config-duplicate-success", name: "valid duplicate config should succeed with the later config value kept"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-authentication -o json", kafkaRestURL), fixture: "kafka/topic/update-topic-config-success-json.golden", name: "config updates with json output"},
		{args: fmt.Sprintf("kafka topic update topic-exist --url %s --config retention.ms=1,compression.type=gzip --no-authentication -o yaml", kafkaRestURL), fixture: "kafka/topic/update-topic-config-success-yaml.golden", name: "config updates with yaml output"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaTopicDescribe() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// Topic name errors
		{args: fmt.Sprintf("kafka topic describe --url %s --no-authentication", kafkaRestURL), contains: "Error: accepts 1 arg(s), received 0", exitCode: 1, name: "<topic> arg missing should lead to error"},
		{args: fmt.Sprintf("kafka topic describe topic-not-exist --url %s --no-authentication", kafkaRestURL), contains: "Error: REST request failed: This server does not host this topic-partition.\n", exitCode: 1, name: "describing a non-existent topic should lead to error"},
		// Success cases
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s --no-authentication", kafkaRestURL), fixture: "kafka/topic/describe-topic-success.golden", name: "topic that exists & correct format arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o human --no-authentication", kafkaRestURL), fixture: "kafka/topic/describe-topic-success.golden", name: "topic that exist & human arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o json --no-authentication", kafkaRestURL), fixture: "kafka/topic/describe-topic-success-json.golden", name: "topic that exist & json arg should lead to success"},
		{args: fmt.Sprintf("kafka topic describe topic-exist --url %s -o yaml --no-authentication", kafkaRestURL), fixture: "kafka/topic/describe-topic-success-yaml.golden", name: "topic that exist & yaml arg should lead to success"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaAcl() {
	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests := []CLITest{
		// error case: bad operation, specified more than one resource type
		{args: fmt.Sprintf("kafka acl list --operation fake --topic Test --consumer-group Group:Test --url %s --no-authentication", kafkaRestURL), name: "bad operation and conflicting resource type errors", fixture: "kafka/acl/list-errors.golden", exitCode: 1},
		// success cases
		{args: fmt.Sprintf("kafka acl list --url %s --no-authentication", kafkaRestURL), name: "acl list output human", fixture: "kafka/acl/list.golden"},
		{args: fmt.Sprintf("kafka acl list -o json --url %s --no-authentication", kafkaRestURL), name: "acl list output json", fixture: "kafka/acl/list-json.golden"},
		{args: fmt.Sprintf("kafka acl list -o yaml --url %s --no-authentication", kafkaRestURL), name: "acl list output yaml", fixture: "kafka/acl/list-yaml.golden"},

		// error case: bad operation, specified more than one resource type, allow/deny not set
		{args: fmt.Sprintf("kafka acl create --principal User:Alice --operation fake --topic Test --consumer-group Group:Test --url %s --no-authentication", kafkaRestURL), name: "bad operation, conflicting resource type, no allow/deny specified errors", fixture: "kafka/acl/create-errors.golden", exitCode: 1},
		// success cases
		{args: fmt.Sprintf("kafka acl create --operation write --cluster-scope --principal User:Alice --allow --url %s --no-authentication", kafkaRestURL), name: "acl create output human", fixture: "kafka/acl/create.golden"},
		{args: fmt.Sprintf("kafka acl create --operation all --cluster-scope --principal User:Alice --allow -o json --url %s --no-authentication", kafkaRestURL), name: "acl create output json", fixture: "kafka/acl/create-json.golden"},
		{args: fmt.Sprintf("kafka acl create --operation all --topic Test --principal User:Alice --allow -o yaml --url %s --no-authentication", kafkaRestURL), name: "acl create output yaml", fixture: "kafka/acl/create-yaml.golden"},

		// error case: bad operation, specified more than one resource type, allow/deny not set
		{args: fmt.Sprintf("kafka acl delete --principal User:Alice --host '*' --operation fake --topic Test --consumer-group Group:Test --url %s --no-authentication", kafkaRestURL), name: "bad operation, conflicting resource type, no allow/deny specified errors", fixture: "kafka/acl/delete-errors.golden", exitCode: 1},
		// success cases
		{args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation read --principal User:Alice --allow --url %s --no-authentication --force", kafkaRestURL), name: "acl delete output human", fixture: "kafka/acl/delete.golden"},
		{args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation read --principal User:Alice --allow --url %s --no-authentication", kafkaRestURL), input: "y\n", name: "acl delete output human", fixture: "kafka/acl/delete-prompt.golden"},
		{args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation read --principal User:Alice --allow -o json --url %s --no-authentication --force", kafkaRestURL), name: "acl delete output json", fixture: "kafka/acl/delete-json.golden"},
		{args: fmt.Sprintf("kafka acl delete --cluster-scope --principal User:Alice --host '*' --operation read --principal User:Alice --allow -o yaml --url %s --no-authentication --force", kafkaRestURL), name: "acl delete output yaml", fixture: "kafka/acl/delete-yaml.golden"},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaQuota() {
	tests := []CLITest{
		// Client Quotas
		{args: "kafka quota create --name clientQuota --description description --ingress 500 --egress 100 --principals sa-1234,sa-5678 --cluster lkc-1234", fixture: "kafka/quota/create.golden"},
		{args: "kafka quota create --name clientQuota --description description --egress 100 --principals sa-1234,sa-5678 --cluster lkc-1234", exitCode: 1, fixture: "kafka/quota/create-no-ingress.golden"},
		{args: "kafka quota create --name clientQuota --ingress 500 --egress 100 --principals \"<default>\" --cluster lkc-1234 -o yaml", fixture: "kafka/quota/create-default-yaml.golden"},
		{args: "kafka quota list --cluster lkc-1234", fixture: "kafka/quota/list.golden"},
		{args: "kafka quota list --cluster lkc-1234 --principal sa-5678 -o json", fixture: "kafka/quota/list-json.golden"},
		{args: "kafka quota list --cluster lkc-1234 -o yaml", fixture: "kafka/quota/list-yaml.golden"},
		{args: "kafka quota describe cq-1234 --cluster lkc-1234", fixture: "kafka/quota/describe.golden"},
		{args: "kafka quota describe cq-1234 --cluster lkc-1234 -o json", fixture: "kafka/quota/describe-json.golden"},
		{args: "kafka quota delete cq-1234 --force", fixture: "kafka/quota/delete.golden"},
		{args: "kafka quota delete cq-1234", input: "quotaName\n", fixture: "kafka/quota/delete-prompt.golden"},
		{args: "kafka quota delete cq-1234 cq-dne", fixture: "kafka/quota/delete-multiple-fail.golden", exitCode: 1},
		{args: "kafka quota delete cq-1234 cq-4321", input: "n\n", fixture: "kafka/quota/delete-multiple-refuse.golden"},
		{args: "kafka quota delete cq-1234 cq-4321", input: "y\n", fixture: "kafka/quota/delete-multiple-success.golden"},
		{args: "kafka quota update cq-1234 --ingress 100 --egress 100 --add-principals sa-4321 --remove-principals sa-1234 --name newName", fixture: "kafka/quota/update.golden"},
		{args: "kafka quota update cq-1234", fixture: "kafka/quota/no-op-update.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaConsumer() {
	tests := []CLITest{
		{args: "kafka consumer list --group consumer-group-1 --cluster lkc-1234", fixture: "kafka/consumer/list.golden"},
		{args: "kafka consumer list --group consumer-group-1 --cluster lkc-1234 -o json", fixture: "kafka/consumer/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	test := CLITest{
		login:   "onprem",
		env:     []string{"CONFLUENT_REST_URL=" + s.TestBackend.GetKafkaRestUrl()},
		args:    "kafka consumer list --group consumer-group-1",
		fixture: "kafka/consumer/list-onprem.golden",
	}

	s.runIntegrationTest(test)
}

func (s *CLITestSuite) TestKafkaConsumerGroup() {
	tests := []CLITest{
		{args: "kafka consumer group list --cluster lkc-1234", fixture: "kafka/consumer/group/list.golden"},
		{args: "kafka consumer group list --cluster lkc-1234 -o json", fixture: "kafka/consumer/group/list-json.golden"},
		{args: "kafka consumer group describe consumer-group-1 --cluster lkc-1234", fixture: "kafka/consumer/group/describe.golden"},
		{args: "kafka consumer group describe consumer-group-1 --cluster lkc-1234 -o json", fixture: "kafka/consumer/group/describe-json.golden"},
		{args: "kafka consumer group describe consumer-group-1234 --cluster lkc-1234", fixture: "kafka/consumer/group/describe-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}

	kafkaRestURL := s.TestBackend.GetKafkaRestUrl()
	tests = []CLITest{
		{args: "kafka consumer group list", fixture: "kafka/consumer/group/list-onprem.golden"},
		{args: "kafka consumer group describe consumer-group-1", fixture: "kafka/consumer/group/describe-onprem.golden"},
		{args: "kafka consumer group describe consumer-group-1234", fixture: "kafka/consumer/group/describe-onprem-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "onprem"
		test.env = []string{"CONFLUENT_REST_URL=" + kafkaRestURL}
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaConsumerGroupLag() {
	tests := []CLITest{
		{args: "kafka consumer group lag describe consumer-group-1 --cluster lkc-describe-dedicated --topic topic-1 --partition 1", fixture: "kafka/consumer/group/lag/describe.golden"},
		{args: "kafka consumer group lag describe consumer-group-1 --cluster lkc-describe-dedicated --topic topic-1 --partition 1 -o json", fixture: "kafka/consumer/group/lag/describe-json.golden"},
		{args: "kafka consumer group lag describe consumer-group-1 --cluster lkc-basic --topic topic-1 --partition 1", fixture: "kafka/consumer/group/lag/basic.golden", exitCode: 1},
		{args: "kafka consumer group lag list consumer-group-1 --cluster lkc-describe-dedicated", fixture: "kafka/consumer/group/lag/list.golden"},
		{args: "kafka consumer group lag list consumer-group-1 --cluster lkc-describe-dedicated -o json", fixture: "kafka/consumer/group/lag/list-json.golden"},
		{args: "kafka consumer group lag list consumer-group-1 --cluster lkc-basic", fixture: "kafka/consumer/group/lag/basic.golden", exitCode: 1},
		{args: "kafka consumer group lag summarize consumer-group-1 --cluster lkc-describe-dedicated", fixture: "kafka/consumer/group/lag/summarize.golden"},
		{args: "kafka consumer group lag summarize consumer-group-1 --cluster lkc-describe-dedicated -o json", fixture: "kafka/consumer/group/lag/summarize-json.golden"},
		{args: "kafka consumer group lag summarize consumer-group-1 --cluster lkc-basic", fixture: "kafka/consumer/group/lag/basic.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaConsumerGroupLag_OnPrem() {
	tests := []CLITest{
		{args: "kafka consumer group lag describe consumer-group-1 --topic topic-1 --partition 1", fixture: "kafka/consumer/group/lag/describe-onprem.golden"},
		{args: "kafka consumer group lag list consumer-group-1", fixture: "kafka/consumer/group/lag/list-onprem.golden"},
		{args: "kafka consumer group lag summarize consumer-group-1", fixture: "kafka/consumer/group/lag/summarize-onprem.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		test.env = []string{"CONFLUENT_REST_URL=" + s.TestBackend.GetKafkaRestUrl()}
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafka_Autocomplete() {
	tests := []CLITest{
		{args: `__complete kafka consumer list --cluster lkc-1234 --group ""`, fixture: "kafka/consumer/list-consumer-group-autocomplete.golden"},
		{args: `__complete kafka consumer group describe --cluster lkc-1234 ""`, fixture: "kafka/consumer/group/autocomplete.golden"},
		{args: `__complete kafka consumer group lag describe --cluster lkc-1234 ""`, fixture: "kafka/consumer/group/lag/describe-autocomplete.golden"},
		{args: `__complete kafka consumer group lag list --cluster lkc-1234 ""`, fixture: "kafka/consumer/group/lag/list-autocomplete.golden"},
		{args: `__complete kafka consumer group lag summarize --cluster lkc-1234 ""`, fixture: "kafka/consumer/group/lag/summarize-autocomplete.golden"},
		{args: `__complete kafka cluster create my-cluster --availability ""`, fixture: "kafka/create-availability-autocomplete.golden"},
		{args: `__complete kafka cluster create my-cluster --type ""`, fixture: "kafka/create-type-autocomplete.golden"},
		{args: `__complete kafka cluster describe ""`, fixture: "kafka/describe-autocomplete.golden"},
		{args: `__complete kafka link delete ""`, fixture: "kafka/link/list-link-delete-autocomplete.golden", useKafka: "lkc-describe-topic"}, // use delete since link has no describe subcommand
		{args: `__complete kafka mirror describe --link link-1 ""`, fixture: "kafka/mirror/describe-autocomplete.golden", useKafka: "lkc-describe-topic"},
		{args: `__complete kafka quota describe ""`, useKafka: "lkc-1234", fixture: "kafka/quota/describe-autocomplete.golden"},
		{args: `__complete kafka topic describe ""`, useKafka: "lkc-describe-topic", fixture: "kafka/topic/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestKafkaClusterCreate_Network() {
	test := CLITest{
		login:   "cloud",
		args:    "kafka cluster create cck-network-test --cloud aws --region us-east-1 --type dedicated --cku 1 --network n-abcde1",
		fixture: "kafka/cluster/cck-network.golden",
	}

	s.runIntegrationTest(test)
}
