package test

import (
	"os"
)

func (s *CLITestSuite) TestCluster() {
	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "true")

	tests := []CLITest{
		{args: "cluster list -o json", fixture: "cluster/confluent-cluster-list-json.golden"},
		{args: "cluster list -o yaml", fixture: "cluster/confluent-cluster-list-yaml.golden"},
		{args: "cluster list", fixture: "cluster/confluent-cluster-list.golden"},
		{args: "connect cluster list", fixture: "cluster/confluent-cluster-list-type-connect.golden"},
		{args: "kafka cluster list", fixture: "cluster/confluent-cluster-list-type-kafka.golden"},
		{args: "ksql cluster list", fixture: "cluster/confluent-cluster-list-type-ksql.golden"},
		{args: "schema-registry cluster list", fixture: "cluster/confluent-cluster-list-type-schema-registry.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}

	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "false")
}

func (s *CLITestSuite) TestClusterRegistry() {
	tests := []CLITest{
		{args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --hosts 10.4.4.4:9004 --protocol PLAIN", fixture: "cluster/confluent-cluster-register-invalid-protocol.golden", wantErrCode: 1},
		{args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --protocol SASL_PLAINTEXT", fixture: "cluster/confluent-cluster-register-missing-hosts.golden", wantErrCode: 1},
		{args: "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol HTTPS"},
		{args: "cluster register --cluster-name theMdsKSQLCluster --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol SASL_PLAINTEXT", fixture: "cluster/confluent-cluster-register-missing-kafka-id.golden", wantErrCode: 1},
		{args: "cluster unregister --cluster-name theMdsKafkaCluster"},
		{args: "cluster unregister", fixture: "cluster/confluent-cluster-unregister-missing-name.golden", wantErrCode: 1},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}
