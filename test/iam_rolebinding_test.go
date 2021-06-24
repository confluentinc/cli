package test

func (s *CLITestSuite) TestCcloudIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			Name: "ccloud iam rolebinding create cloud-cluster",
			Args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			Name: "ccloud iam rolebinding create cloud-cluster",
			Args: "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			Name:        "ccloud iam rolebinding create, invalid use case: missing cloud-cluster",
			Args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			Args:        "iam rolebinding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding create, invalid use case: missing environment",
			Args:        "iam rolebinding create --principal User:u-11aaa --role EnvironmentAdmin",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			WantErrCode: 1,
		},
		{
			Name: "ccloud iam rolebinding delete cluster-Name",
			Args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa",
		},
		{
			Name: "ccloud iam rolebinding delete cluster-Name",
			Args: "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
		},
		{
			Name:        "ccloud iam rolebinding delete, invalid use case: missing cloud-cluster",
			Args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-cloud-cluster.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			Args:        "iam rolebinding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding delete, invalid use case: missing environment",
			Args:        "iam rolebinding delete --principal User:u-11aaa --role EnvironmentAdmin",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-missing-environment.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding delete cluster-Name, invalid use case: missing role",
			Args:        "iam rolebinding delete --principal User:u-11aaa --current-env --cloud-cluster lkc-1111aaa",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-delete-missing-role.golden",
			WantErrCode: 1,
		},
		{
			Name:		"ccloud iam rolebinding create with email as principal",
			Args:		"iam rolebinding create --principal User:u-11aaa@confluent.io --role CloudClusterAdmin --current-env --cloud-cluster lkc-1111aaa",
			Fixture: 	"iam-rolebinding/ccloud-iam-rolebinding-create-with-email.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}

func (s *CLITestSuite) TestConfluentIAMRoleBindingCRUD() {
	tests := []CLITest{
		{
			Name:    "confluent iam rolebinding create --help",
			Args:    "iam rolebinding create --help",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-create-help.golden",
		},
		{
			Name:    "confluent iam rolebinding create cluster-Name",
			Args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-Name theMdsConnectCluster",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-create-cluster-Name.golden",
		},
		{
			Name:    "confluent iam rolebinding create cluster-id",
			Args:    "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-create-cluster-id.golden",
		},
		{
			Name:        "confluent iam rolebinding create, invalid use case: cluster-Name & kafka-cluster-id specified",
			Args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-Name theMdsConnectCluster",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-Name-and-id-error.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding create, invalid use case: cluster-Name & ksql-cluster-id specified",
			Args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --cluster-Name theMdsConnectCluster",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-Name-and-id-error.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding create, invalid use case: missing cluster-Name or cluster-id",
			Args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-Name-or-id.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding create, invalid use case: missing kafka-cluster-id",
			Args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-Name",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding create, invalid use case: multiple non kafka id",
			Args:        "iam rolebinding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			WantErrCode: 1,
		},
		{
			Name:    "confluent iam rolebinding delete --help",
			Args:    "iam rolebinding delete --help",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-delete-help.golden",
		},
		{
			Name:    "confluent iam rolebinding delete cluster-Name",
			Args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-Name theMdsConnectCluster",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-delete-cluster-Name.golden",
		},
		{
			Name:    "confluent iam rolebinding delete cluster-id",
			Args:    "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-delete-cluster-id.golden",
		},
		{
			Name:        "confluent iam rolebinding delete, invalid use case: cluster-Name & kafka-cluster-id specified",
			Args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID --cluster-Name theMdsConnectCluster",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-Name-and-id-error.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding delete, invalid use case: cluster-Name & ksql-cluster-id specified",
			Args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --cluster-Name theMdsConnectCluster",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-Name-and-id-error.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding delete, invalid use case: missing cluster-Name or cluster-id",
			Args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-Name-or-id.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding delete, invalid use case: missing  kafka-cluster-id",
			Args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksql-Name",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-missing-kafka-cluster-id.golden",
			WantErrCode: 1,
		},
		{
			Name:        "confluent iam rolebinding delete, invalid use case: multiple non kafka id",
			Args:        "iam rolebinding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster-id ksqlName --connect-cluster-id connectID --kafka-cluster-id kafka-GUID",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-multiple-non-kafka-id.golden",
			WantErrCode: 1,
		},
		{
			Name:    "confluent iam rolebinding create principal with @",
			Args:    "iam rolebinding create --principal User:bob@Kafka --role DeveloperRead --resource Topic:connect-configs --kafka-cluster-id kafka-GUID",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-create-cluster-id-at.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestConfluentIAMRolebindingList() {
	tests := []CLITest{
		{
			Name:    "confluent iam rolebinding list --help",
			Args:    "iam rolebinding list --help",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-help.golden",
		},
		{
			Name:        "confluent iam rolebinding list, no principal nor role",
			Args:        "iam rolebinding list --kafka-cluster-id CID",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-list-no-principal-nor-role.golden",
			WantErrCode: 1,
		},
		{
			Args:        "iam rolebinding list --kafka-cluster-id CID --principal frodo",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-list-principal-format-error.golden",
			WantErrCode: 1,
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user.golden",
		},
		{
			Args:    "iam rolebinding list --cluster-Name kafka --principal User:frodo",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user.golden",
		},
		{
			Args:        "iam rolebinding list --cluster-Name kafka --kafka-cluster-id CID --principal User:frodo",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-Name-and-id-error.golden",
			WantErrCode: 1,
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o json",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-json.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o yaml",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-yaml.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperWrite",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-resources-from-multiple-groups.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SecurityAdmin",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-cluster-resource.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o json",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches-json.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o yaml",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-user-and-role-with-no-matches-yaml.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperRead",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-multiple-resources.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperWrite",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-one-resource.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role SecurityAdmin",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-group-and-role-with-no-matches.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o json",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-json.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o yaml",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-yaml.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-bindings-to-multiple-groups.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role SecurityAdmin",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-on-cluster-bound-to-user.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role SystemAdmin",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-with-no-matches.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:food",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-exact-match.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:shire-parties",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-no-match.golden",
		},
		{
			Args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite --resource Topic:shire-parties",
			Fixture: "iam-rolebinding/confluent-iam-rolebinding-list-role-and-resource-with-prefix-match.golden",
		},
		{
			Name:        "confluent iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			Args:        "iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			Fixture:     "iam-rolebinding/confluent-iam-rolebinding-list-failure-help.golden",
			WantErrCode: 1,
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestCcloudIAMRolebindingList() {
	tests := []CLITest{
		{
			Name:        "ccloud iam rolebinding list, no principal nor role",
			Args:        "iam rolebinding list",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			WantErrCode: 1,
		},
		{
			Name:        "ccloud iam rolebinding list, no principal nor role",
			Args:        "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			WantErrCode: 1,
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-1.golden",
		},
		{
			Args:    "iam rolebinding list --current-env --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-1.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-22bbb",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-2.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-33ccc",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-3.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-44ddd",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-4.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role OrganizationAdmin",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-orgadmin.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role EnvironmentAdmin",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-envadmin.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin-yaml.golden",
		},
		{
			Args:    "iam rolebinding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-user-clusteradmin-json.golden",
		},
		{
			Name:        "ccloud iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			Args:        "iam rolebinding list --principal User:u-41dxz3 --cluster pantsCluster",
			Fixture:     "iam-rolebinding/ccloud-iam-rolebinding-list-failure-help.golden",
			WantErrCode: 1,
		},
		{
			Name:    "ccloud iam rolebinding list --help",
			Args:    "iam rolebinding list --help",
			Fixture: "iam-rolebinding/ccloud-iam-rolebinding-list-help.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
