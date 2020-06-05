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
			fixture:     "confluent-iam-rolebinding-name-and-id.golden",
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

func (s *CLITestSuite) Test_Confluent_Iam_Role_List() {
	tests := []CLITest{
		{
			name:        "confluent iam role list",
			args:        "iam role list",
			fixture:     "confluent-iam-role-list.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam role list json",
			args:        "iam role list -o json",
			fixture:     "confluent-iam-role-list-json.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam role list yaml",
			args:        "iam role list -o yaml",
			fixture:     "confluent-iam-role-list-yaml.golden",
			login:       "default",
			wantErrCode: 0,
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}

func (s *CLITestSuite) Test_Confluent_Iam_Role_Describe() {
	tests := []CLITest{
		{
			name:        "confluent iam role describe",
			args:        "iam role describe DeveloperRead",
			fixture:     "confluent-iam-role-describe.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam role describe json",
			args:        "iam role describe DeveloperRead -o json",
			fixture:     "confluent-iam-role-describe-json.golden",
			login:       "default",
			wantErrCode: 0,
		},
		{
			name:        "confluent iam role describe yaml",
			args:        "iam role describe DeveloperRead -o yaml",
			fixture:     "confluent-iam-role-describe-yaml.golden",
			login:       "default",
			wantErrCode: 0,
		},
	}
	for _, tt := range tests {
		s.runConfluentTest(tt, serveMds(s.T()).URL)
	}
}
