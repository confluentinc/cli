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
			args:    "cluster list --type ksql-cluster",
			fixture: "confluent-cluster-list-type-ksql.golden",
			login:   "default",
		},
		{
			args:    "cluster list --type unknown",
			fixture: "confluent-cluster-list-type-unknown.golden",
			login:   "default",
			wantErrCode: 1,
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