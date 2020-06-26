package test

func (s *CLITestSuite) Test_Cluster_Registry() {

	tests := []CLITest{
		{
			args:        "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol HTTPS",
			login:       "default",
			wantErrCode: 0,
		},
		{
			args:        "cluster register --cluster-name theMdsKSQLCluster --ksql-cluster-id ksql-name --hosts 10.4.4.4:9004 --protocol SASL_PLAINTEXT",
			fixture:     "confluent-cluster-register-missing-kafka-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:        "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --protocol SASL_PLAINTEXT",
			fixture:     "confluent-cluster-register-missing-hosts.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:        "cluster register --cluster-name theMdsKSQLCluster --kafka-cluster-id kafka-GUID --ksql-cluster-id  ksql-name --hosts 10.4.4.4:9004 --protocol PLAIN",
			fixture:     "confluent-cluster-register-invalid-protocol.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:        "cluster unregister --cluster-name theMdsKafkaCluster",
			login:       "default",
			wantErrCode: 0,
		},
		{
			args:        "cluster unregister",
			fixture:     "confluent-cluster-unregister-missing-name.golden",
			login:       "default",
			wantErrCode: 1,
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}
