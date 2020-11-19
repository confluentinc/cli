package test

func (s *CLITestSuite) TestCcloudIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			name: "ccloud iam rolebinding create cloud-cluster",
			args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			name: "ccloud iam rolebinding create cloud-cluster",
			args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name: "ccloud iam rolebinding delete cluster-name",
			args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			name: "ccloud iam rolebinding delete cluster-name",
			args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			wantErrCode: 1,
		},
	}

	kafkaURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaURL).URL

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt, loginURL)
	}
}

func (s *CLITestSuite) TestConfluentIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			name:    "confluent iam rolebinding create cluster-name",
			args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-create-cluster-name.golden",
		},
		{
			name:    "confluent iam rolebinding create cluster-id",
			args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-create-cluster-id.golden",
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-name-or-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: missing kafka-cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding create, invalid use case: multiple non kafka id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			wantErrCode: 1,
		},
		{
			name:    "confluent iam rolebinding delete cluster-name",
			args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-delete-cluster-name.golden",
		},
		{
			name:    "confluent iam rolebinding delete cluster-id",
			args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-delete-cluster-id.golden",
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-name-or-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: missing  kafka-cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "confluent iam rolebinding delete, invalid use case: multiple non kafka id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			wantErrCode: 1,
		},
	}

	loginURL := serveMds(s.T()).URL

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt, loginURL)
	}
}

func (s *CLITestSuite) TestConfluentIAMRolebindingList() {
	tests := []CLITest{
		{
			name:        "confluent iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --kafka-cluster-id CID",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			args:        "iam rolebinding list --kafka-cluster-id CID --principal frodo",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-list-principal-format-error.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user.golden",
		},
		{
			args:    "iam rolebinding list --cluster-name kafka --principal User:frodo",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user.golden",
		},
		{
			args:        "iam rolebinding list --cluster-name kafka --kafka-cluster-id CID --principal User:frodo",
			fixture:     "iam-rolebinding/confluent-iam-rolebinding-name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o json",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o yaml",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperWrite",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-resources-from-multiple-groups.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SecurityAdmin",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-cluster-resource.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o json",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o yaml",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperRead",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-multiple-resources.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperWrite",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-one-resource.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role SecurityAdmin",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o json",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o yaml",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-bindings-to-multiple-groups.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SecurityAdmin",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-on-cluster-bound-to-user.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SystemAdmin",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:food",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-exact-match.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:shire-parties",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-no-match.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite --resource Topic:shire-parties",
			fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-prefix-match.golden",
		},
	}

	loginURL := serveMds(s.T()).URL

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt, loginURL)
	}
}

func (s *CLITestSuite) TestCcloudIAMRolebindingList() {
	tests := []CLITest{
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-1.golden",
		},
		{
			args:    "iam rolebinding list --current-env --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-1.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-22bbb",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-2.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-33ccc",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-3.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-44ddd",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-4.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role OrganizationAdmin",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-orgadmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role EnvironmentAdmin",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-envadmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin-yaml.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json",
			fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin-json.golden",
		},
	}

	kafkaURL := serveKafkaAPI(s.T()).URL
	loginURL := serve(s.T(), kafkaURL).URL

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt, loginURL)
	}
}
