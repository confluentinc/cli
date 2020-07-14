package test

func (s *CLITestSuite) Test_Ccloud_Iam_Role_Binding_CRUD() {
	tests := []CLITest{
		{
			name:  "ccloud iam rolebinding create cloud-cluster",
			args:  "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --environment current --cloud-cluster lkc-1111aaa",
			login: "default",
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "ccloud-iam-rolebinding-missing-environment.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "ccloud-iam-rolebinding-missing-environment.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:  "ccloud iam rolebinding delete cluster-name",
			args:  "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --environment current --cloud-cluster lkc-1111aaa",
			login: "default",
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "ccloud-iam-rolebinding-missing-environment.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "ccloud-iam-rolebinding-missing-environment.golden",
			login:       "default",
			wantErrCode: 1,
		},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) Test_Confluent_Iam_Role_Binding_CRUD() {
	tests := []CLITest{
		{
			name:        "confluent iam rolebinding create cluster-name",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-create-cluster-name.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam rolebinding create cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture:     "confluent-iam-create-cluster-id.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-rolebinding-name-and-id-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-rolebinding-name-and-id-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "confluent-iam-rolebinding-missing-name-or-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: missing kafka-cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: multiple non kafka id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete cluster-name",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-delete-cluster-name.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam rolebinding delete cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture:     "confluent-iam-delete-cluster-id.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-rolebinding-name-and-id-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "confluent-iam-rolebinding-name-and-id-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "confluent-iam-rolebinding-missing-name-or-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: missing  kafka-cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: multiple non kafka id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			login:       "default",
			wantErrCode: 1,
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}
