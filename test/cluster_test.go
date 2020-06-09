package test

import (
	"os"
)

func (s *CLITestSuite) Test_Cluster() {
	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "true")

	tests := []CLITest{
		{
			args:    "cluster list",
			fixture: "confluent-cluster-list.golden",
			login:   "default",
		},
		{
			args:    "ksql cluster list",
			fixture: "confluent-cluster-list-type-ksql.golden",
			login:   "default",
		},
		{
			args:    "kafka cluster list",
			fixture: "confluent-cluster-list-type-kafka.golden",
			login:   "default",
		},
		{
			args:    "schema-registry cluster list",
			fixture: "confluent-cluster-list-type-schema-registry.golden",
			login:   "default",
		},
		{
			args:    "connect cluster list",
			fixture: "confluent-cluster-list-type-connect.golden",
			login:   "default",
		},
		{
			args:    "cluster list -o json",
			fixture: "confluent-cluster-list-json.golden",
			login:   "default",
		},
		{
			args:    "cluster list -o yaml",
			fixture: "confluent-cluster-list-yaml.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}

	_ = os.Setenv("XX_FLAG_CLUSTER_REGISTRY_ENABLE", "false")
}
