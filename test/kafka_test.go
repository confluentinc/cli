package test

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
		{args: "kafka acl list --cluster lkc-acls", fixture: "kafka/rp-kafka-acls-list.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBED --topic 'test-topic'", fixture: "kafka/kafka-acls-invalid-operation.golden", wantErrCode: 1},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl create --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'"},
		{args: "kafka acl delete --cluster lkc-acls --allow --service-account 7272 --operation READ --operation DESCRIBE --topic 'test-topic'", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},

		{args: "kafka link list --cluster lkc-links", fixture: "kafka/kafka20.golden", wantErrCode: 0},
		{args: "kafka link list --cluster lkc-links -o json", fixture: "kafka/kafka21.golden", wantErrCode: 0},
		{args: "kafka link list --cluster lkc-links -o yaml", fixture: "kafka/kafka22.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link", fixture: "kafka/kafka23.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link -o json", fixture: "kafka/kafka24.golden", wantErrCode: 0},
		{args: "kafka link describe --cluster lkc-links my-link -o yaml", fixture: "kafka/kafka25.golden", wantErrCode: 0},

		{args: "kafka topic mirror stop test-topic", login: "default", useKafka: "lkc-topics", authKafka: "true"},
		{args: "kafka topic mirror stop not-found", login: "default", useKafka: "lkc-topics", authKafka: "true", fixture: "kafka/mirror-topic-not-found.golden", wantErrCode: 1},
		{args: "kafka topic mirror bad test-topic", login: "default", useKafka: "lkc-topics", authKafka: "true", fixture: "kafka/mirror-invalid.golden", wantErrCode: 1},

		{args: "kafka topic list", login: "default", useKafka: "lkc-topics", fixture: "kafka/topic-list.golden"},
		{args: "kafka topic list --cluster lkc-topics", login: "default", fixture: "kafka/topic-list.golden"},
		{args: "kafka topic list --cluster lkc-topics", fixture: "kafka/rp-topic-list.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka topic list", login: "default", useKafka: "lkc-no-topics", fixture: "kafka/topic-list-empty.golden"},
		{args: "kafka topic list", login: "default", useKafka: "lkc-not-ready", fixture: "kafka/cluster-not-ready.golden", wantErrCode: 1},

		{args: "kafka topic create", login: "default", useKafka: "lkc-create-topic", fixture: "kafka/topic-create.golden", wantErrCode: 1},
		{args: "kafka topic create topic1", login: "default", useKafka: "lkc-create-topic", fixture: "kafka/topic-create-success.golden"},
		{args: "kafka topic create topic1", useKafka: "lkc-create-topic", fixture: "kafka/topic-create-success.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka topic create dupTopic", login: "default", useKafka: "lkc-create-topic", fixture: "kafka/topic-create-dup-topic.golden", wantErrCode: 1},

		{args: "kafka topic describe", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-describe.golden", wantErrCode: 1},
		{args: "kafka topic describe topic1", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-describe-success.golden"},
		{args: "kafka topic describe topic1", useKafka: "lkc-describe-topic", fixture: "kafka/rp-topic-describe-success.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka topic describe topic1 --output json", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-describe-json-success.golden"},
		{args: "kafka topic describe topic1 --cluster lkc-create-topic", login: "default", fixture: "kafka/topic-describe-not-found.golden", wantErrCode: 1},
		{args: "kafka topic describe topic2", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic2-describe-not-found.golden", wantErrCode: 1},

		{args: "kafka topic delete", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete.golden", wantErrCode: 1},
		{args: "kafka topic delete topic1", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete-success.golden"},
		{args: "kafka topic delete topic1", useKafka: "lkc-delete-topic", fixture: "kafka/topic-delete-success.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
		{args: "kafka topic delete topic1 --cluster lkc-create-topic", login: "default", fixture: "kafka/topic-delete-not-found.golden", wantErrCode: 1},
		{args: "kafka topic delete topic2", login: "default", useKafka: "lkc-delete-topic", fixture: "kafka/topic2-delete-not-found.golden", wantErrCode: 1},

		{args: "kafka topic update topic1 --config=\"testConfig=valueUpdate\"", login: "default", useKafka: "lkc-describe-topic", fixture: "kafka/topic-update-success.golden"},
		{args: "kafka topic update topic1 --config=\"testConfig=valueUpdate\"", useKafka: "lkc-describe-topic", fixture: "kafka/topic-update-success.golden", env: []string{"XX_CCLOUD_USE_KAFKA_REST=true"}},
	}

	resetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.login = "default"
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}
