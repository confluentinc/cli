package test

func (s *CLITestSuite) Test_Confluent_Iam_Rolebinding_List() {
	tests := []CLITest{
		{
			name:        "confluent iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --kafka-cluster-id CID",
			fixture:     "confluent-iam-rolebinding-list-no-principal-nor-role.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:        "iam rolebinding list --kafka-cluster-id CID --principal frodo",
			fixture:     "confluent-iam-rolebinding-list-principal-format-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo",
			fixture: "confluent-iam-rolebinding-list-user.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --cluster-name kafka --principal User:frodo",
			fixture: "confluent-iam-rolebinding-list-user.golden",
			login:   "default",
		},
		{
			args:        "iam rolebinding list --cluster-name kafka  --kafka-cluster-id CID --principal User:frodo",
			fixture:     "confluent-iam-rolebinding-name-and-id-error.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o json",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-json.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperRead -o yaml",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-multiple-resources-from-one-group-yaml.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-resources-from-multiple-groups.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-cluster-resource.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o json",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-no-matches-json.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal User:frodo --role SystemAdmin -o yaml",
			fixture: "confluent-iam-rolebinding-list-user-and-role-with-no-matches-yaml.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-multiple-resources.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-one-resource.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --principal Group:hobbits --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-group-and-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead",
			fixture: "confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o json",
			fixture: "confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-json.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead -o yaml",
			fixture: "confluent-iam-rolebinding-list-role-with-multiple-bindings-to-one-group-yaml.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite",
			fixture: "confluent-iam-rolebinding-list-role-with-bindings-to-multiple-groups.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SecurityAdmin",
			fixture: "confluent-iam-rolebinding-list-role-on-cluster-bound-to-user.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role SystemAdmin",
			fixture: "confluent-iam-rolebinding-list-role-with-no-matches.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:food",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-exact-match.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperRead --resource Topic:shire-parties",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-no-match.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --kafka-cluster-id CID --role DeveloperWrite --resource Topic:shire-parties",
			fixture: "confluent-iam-rolebinding-list-role-and-resource-with-prefix-match.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) Test_Ccloud_Iam_Rolebinding_List() {
	tests := []CLITest{
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list",
			fixture:     "ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:        "ccloud iam rolebinding list, no principal nor role",
			args:        "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa",
			fixture:     "ccloud-iam-rolebinding-list-no-principal-nor-role.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --principal User:u-11aaa",
			fixture: "ccloud-iam-rolebinding-list-user-1.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --principal User:u-22bbb",
			fixture: "ccloud-iam-rolebinding-list-user-2.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --principal User:u-33ccc",
			fixture: "ccloud-iam-rolebinding-list-user-3.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --principal User:u-44ddd",
			fixture: "ccloud-iam-rolebinding-list-user-4.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --role OrganizationAdmin",
			fixture: "ccloud-iam-rolebinding-list-user-orgadmin.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --role EnvironmentAdmin",
			fixture: "ccloud-iam-rolebinding-list-user-envadmin.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --role CloudClusterAdmin",
			fixture: "ccloud-iam-rolebinding-list-user-clusteradmin.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml",
			fixture: "ccloud-iam-rolebinding-list-user-clusteradmin-yaml.golden",
			login:   "default",
		},
		{
			args:    "iam rolebinding list --environment current --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json",
			fixture: "ccloud-iam-rolebinding-list-user-clusteradmin-json.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) Test_Confluent_Iam_Role_List() {
	tests := []CLITest{
		{
			name:    "confluent iam role list",
			args:    "iam role list",
			fixture: "confluent-iam-role-list.golden",
			login:   "default",
		},
		{
			name:    "confluent iam role list json",
			args:    "iam role list -o json",
			fixture: "confluent-iam-role-list-json.golden",
			login:   "default",
		},
		{
			name:    "confluent iam role list yaml",
			args:    "iam role list -o yaml",
			fixture: "confluent-iam-role-list-yaml.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) Test_Ccloud_Iam_Role_List() {
	tests := []CLITest{
		{
			name:    "ccloud iam role list",
			args:    "iam role list",
			fixture: "ccloud-iam-role-list.golden",
			login:   "default",
		},
		{
			name:    "ccloud iam role list json",
			args:    "iam role list -o json",
			fixture: "ccloud-iam-role-list-json.golden",
			login:   "default",
		},
		{
			name:    "ccloud iam role list yaml",
			args:    "iam role list -o yaml",
			fixture: "ccloud-iam-role-list-yaml.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}

func (s *CLITestSuite) Test_Confluent_Iam_Role_Describe() {
	tests := []CLITest{
		{
			name:    "confluent iam role describe",
			args:    "iam role describe DeveloperRead",
			fixture: "confluent-iam-role-describe.golden",
			login:   "default",
		},
		{
			name:    "confluent iam role describe json",
			args:    "iam role describe DeveloperRead -o json",
			fixture: "confluent-iam-role-describe-json.golden",
			login:   "default",
		},
		{
			name:    "confluent iam role describe yaml",
			args:    "iam role describe DeveloperRead -o yaml",
			fixture: "confluent-iam-role-describe-yaml.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) Test_Ccloud_Iam_Role_Describe() {
	tests := []CLITest{
		{
			name:    "ccloud iam role describe",
			args:    "iam role describe CloudClusterAdmin",
			fixture: "ccloud-iam-role-describe.golden",
			login:   "default",
		},
		{
			name:        "ccloud iam role describe invalid role",
			args:        "iam role describe InvalidRole",
			fixture:     "ccloud-iam-role-describe-invalid-role.golden",
			login:       "default",
			wantErrCode: 1,
		},
		{
			name:    "ccloud iam role describe json",
			args:    "iam role describe CloudClusterAdmin -o json",
			fixture: "ccloud-iam-role-describe-json.golden",
			login:   "default",
		},
		{
			name:    "ccloud iam role describe yaml",
			args:    "iam role describe CloudClusterAdmin -o yaml",
			fixture: "ccloud-iam-role-describe-yaml.golden",
			login:   "default",
		},
	}
	for _, tt := range tests {
		kafkaAPIURL := serveKafkaAPI(s.T()).URL
		s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL)
	}
}
