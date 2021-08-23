package test

func (s *CLITestSuite) TestCloudIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			name: "iam rolebinding create cloud-cluster",
			args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			name: "iam rolebinding create cloud-cluster",
			args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			name:        "iam rolebinding create, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "iam-rolebinding/missing-cloud-cluster.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: missing environment",
			args:        "iam rolebinding create --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "iam-rolebinding/missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name: "iam rolebinding delete cluster-name",
			args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			name: "iam rolebinding delete cluster-name",
			args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			name:        "iam rolebinding delete, invalid use case: missing cloud-cluster",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin",
			fixture:     "iam-rolebinding/missing-cloud-cluster.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: missing environment",
			args:        "iam rolebinding delete --principal User:u-11aaa --role EnvironmentAdmin",
			fixture:     "iam-rolebinding/missing-environment.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete cluster-name, invalid use case: missing role",
			args:        "iam rolebinding delete --principal User:u-11aaa --current-env --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/delete-missing-role.golden",
			wantErrCode: 1,
		},
		{
			name:    "iam rolebinding create with email as principal",
			args:    "iam rolebinding create --principal User:u-11aaa@confluent.io --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
			fixture: "iam-rolebinding/create-with-email.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCloudTest(tt)
	}
}

func (s *CLITestSuite) TestOnPremIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			args:    "iam rolebinding create --help",
			fixture: "iam-rolebinding/create-help.golden",
		},
		{
			name:    "iam rolebinding create cluster-name",
			args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture: "iam-rolebinding/create-cluster-name.golden",
		},
		{
			name:    "iam rolebinding create cluster-id",
			args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture: "iam-rolebinding/create-cluster-id.golden",
		},
		{
			name:        "iam rolebinding create, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "iam-rolebinding/missing-name-or-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: missing kafka-cluster-id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "iam-rolebinding/missing-kafka-cluster-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding create, invalid use case: multiple non kafka id",
			args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "iam-rolebinding/multiple-non-kafka-id.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding delete --help",
			fixture: "iam-rolebinding/delete-help.golden",
		},
		{
			name:    "iam rolebinding delete cluster-name",
			args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster",
			fixture: "iam-rolebinding/delete-cluster-name.golden",
		},
		{
			name:    "iam rolebinding delete cluster-id",
			args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture: "iam-rolebinding/delete-cluster-id.golden",
		},
		{
			name:        "iam rolebinding delete, invalid use case: cluster-name & kafka-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: cluster-name & ksql-cluster-id specified",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlname --cluster-name theMdsConnectCluster",
			fixture:     "iam-rolebinding/name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: missing cluster-name or cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			fixture:     "iam-rolebinding/missing-name-or-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: missing  kafka-cluster-id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-name",
			fixture:     "iam-rolebinding/missing-kafka-cluster-id.golden",
			wantErrCode: 1,
		},
		{
			name:        "iam rolebinding delete, invalid use case: multiple non kafka id",
			args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			fixture:     "iam-rolebinding/multiple-non-kafka-id.golden",
			wantErrCode: 1,
		},
		{
			name:    "iam rolebinding create principal with @",
			args:    "iam rolebinding create --principal User:bob@Kafka --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			fixture: "iam-rolebinding/create-cluster-id-at.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runOnPremTest(tt)
	}
}

func (s *CLITestSuite) TestOnPremIAMRolebindingList() {
	tests := []CLITest{
		{
			args:    "iam rolebinding list --help",
			fixture: "iam-rolebinding/list-help-onprem.golden",
		},
		{
			name:        "iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --kafka-cluster-id CID",
			fixture:     "iam-rolebinding/list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			args:        "iam rolebinding list --kafka-cluster-id CID --principal frodo",
			fixture:     "iam-rolebinding/list-principal-format-error.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo",
			fixture: "iam-rolebinding/list-user.golden",
		},
		{
			args:    "iam rolebinding list --cluster-name kafka --principal User:frodo",
			fixture: "iam-rolebinding/list-user.golden",
		},
		{
			args:        "iam rolebinding list --cluster-name kafka --kafka-cluster-id CID --principal User:frodo",
			fixture:     "iam-rolebinding/name-and-id-error.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead",
			fixture: "iam-rolebinding/list-user-and-role-with-multiple-resources-from-one-group.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o json",
			fixture: "iam-rolebinding/list-user-and-role-with-multiple-resources-from-one-group-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o yaml",
			fixture: "iam-rolebinding/list-user-and-role-with-multiple-resources-from-one-group-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperWrite",
			fixture: "iam-rolebinding/list-user-and-role-with-resources-from-multiple-groups.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SecurityAdmin",
			fixture: "iam-rolebinding/list-user-and-role-with-cluster-resource.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin",
			fixture: "iam-rolebinding/list-user-and-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o json",
			fixture: "iam-rolebinding/list-user-and-role-with-no-matches-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o yaml",
			fixture: "iam-rolebinding/list-user-and-role-with-no-matches-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperRead",
			fixture: "iam-rolebinding/list-group-and-role-with-multiple-resources.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperWrite",
			fixture: "iam-rolebinding/list-group-and-role-with-one-resource.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role SecurityAdmin",
			fixture: "iam-rolebinding/list-group-and-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead",
			fixture: "iam-rolebinding/list-role-with-multiple-bindings-to-one-group.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o json",
			fixture: "iam-rolebinding/list-role-with-multiple-bindings-to-one-group-json.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o yaml",
			fixture: "iam-rolebinding/list-role-with-multiple-bindings-to-one-group-yaml.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite",
			fixture: "iam-rolebinding/list-role-with-bindings-to-multiple-groups.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SecurityAdmin",
			fixture: "iam-rolebinding/list-role-on-cluster-bound-to-user.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SystemAdmin",
			fixture: "iam-rolebinding/list-role-with-no-matches.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:food",
			fixture: "iam-rolebinding/list-role-and-resource-with-exact-match.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:shire-parties",
			fixture: "iam-rolebinding/list-role-and-resource-with-no-match.golden",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite --resource Topic:shire-parties",
			fixture: "iam-rolebinding/list-role-and-resource-with-prefix-match.golden",
		},
		{
			args:        "iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			fixture:     "iam-rolebinding/list-failure-help-onprem.golden",
			wantErrCode: 1,
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runOnPremTest(tt)
	}
}

func (s *CLITestSuite) TestCcloudIAMRolebindingList() {
	tests := []CLITest{
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list",
			fixture:     "iam-rolebinding/list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa",
			fixture:     "iam-rolebinding/list-no-principal-nor-role.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			fixture: "iam-rolebinding/list-user-1.golden",
		},
		{
			args:    "iam rolebinding list --current-env --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			fixture: "iam-rolebinding/list-user-1.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-22bbb",
			fixture: "iam-rolebinding/list-user-2.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-33ccc",
			fixture: "iam-rolebinding/list-user-3.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-44ddd",
			fixture: "iam-rolebinding/list-user-4.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role OrganizationAdmin",
			fixture: "iam-rolebinding/list-user-orgadmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role EnvironmentAdmin",
			fixture: "iam-rolebinding/list-user-envadmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin",
			fixture: "iam-rolebinding/list-user-clusteradmin.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml",
			fixture: "iam-rolebinding/list-user-clusteradmin-yaml.golden",
		},
		{
			args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json",
			fixture: "iam-rolebinding/list-user-clusteradmin-json.golden",
		},
		{
			args:        "iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			fixture:     "iam-rolebinding/list-failure-help-cloud.golden",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --help",
			fixture: "iam-rolebinding/list-help-cloud.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCloudTest(tt)
	}
}
