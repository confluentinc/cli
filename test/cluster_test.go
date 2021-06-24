package test

import (
	"os"
)

func (s *CLITestSuite) TestCluster() {
	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "true")

	tests := []CLITest{
		{Args: "cluster list --help", Fixture: "cluster/confluent-cluster-list-help.golden"},
		{Args: "cluster list -o json", Fixture: "cluster/confluent-cluster-list-json.golden"},
		{Args: "cluster list -o yaml", Fixture: "cluster/confluent-cluster-list-yaml.golden"},
		{Args: "cluster list", Fixture: "cluster/confluent-cluster-list.golden"},
		{Args: "connect cluster list --help", Fixture: "cluster/confluent-cluster-connect-list-help.golden"},
		{Args: "connect cluster list", Fixture: "cluster/confluent-cluster-list-type-connect.golden"},
		{Args: "kafka cluster list --help", Fixture: "cluster/confluent-cluster-kafka-list-help.golden"},
		{Args: "kafka cluster list", Fixture: "cluster/confluent-cluster-list-type-kafka.golden"},
		{Args: "ksql cluster list --help", Fixture: "cluster/confluent-cluster-ksql-list-help.golden"},
		{Args: "ksql cluster list", Fixture: "cluster/confluent-cluster-list-type-ksql.golden"},
		{Args: "schema-registry cluster list --help", Fixture: "cluster/confluent-cluster-schema-registry-list-help.golden"},
		{Args: "schema-registry cluster list", Fixture: "cluster/confluent-cluster-list-type-schema-registry.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}

	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "false")
}

func (s *CLITestSuite) TestClusterRegistry() {
	tests := []CLITest{
		{Args: "cluster register --help", Fixture: "cluster/confluent-cluster-register-list-help.golden"},
		{Args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --hosts 10.4.4.4:9004 --protocol PLAIN", Fixture: "cluster/confluent-cluster-register-invalid-protocol.golden", WantErrCode: 1},
		{Args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --protocol SASL_PLAINTEXT", Fixture: "cluster/confluent-cluster-register-missing-hosts.golden", WantErrCode: 1},
		{Args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol HTTPS"},
		{Args: "cluster register --cluster-name theMdsKSQLCluster --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol SASL_PLAINTEXT", Fixture: "cluster/confluent-cluster-register-missing-kafka-id.golden", WantErrCode: 1},
		{Args: "cluster unregister --help", Fixture: "cluster/confluent-cluster-unregister-list-help.golden"},
		{Args: "cluster unregister --cluster-name theMdsKafkaCluster"},
		{Args: "cluster unregister", Fixture: "cluster/confluent-cluster-unregister-missing-name.golden", WantErrCode: 1},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}
