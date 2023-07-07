package test

func (s *CLITestSuite) TestIamRbacRole_OnPrem() {
	tests := []CLITest{
		{args: "iam rbac role describe DeveloperRead -o json", fixture: "iam/rbac/role/describe-json-onprem.golden"},
		{args: "iam rbac role describe DeveloperRead -o yaml", fixture: "iam/rbac/role/describe-yaml-onprem.golden"},
		{args: "iam rbac role describe DeveloperRead", fixture: "iam/rbac/role/describe-onprem.golden"},
		{args: "iam rbac role list -o json", fixture: "iam/rbac/role/list-json-onprem.golden"},
		{args: "iam rbac role list -o yaml", fixture: "iam/rbac/role/list-yaml-onprem.golden"},
		{args: "iam rbac role list", fixture: "iam/rbac/role/list-onprem.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRole_Cloud() {
	tests := []CLITest{
		{args: "iam rbac role describe CloudClusterAdmin -o json", fixture: "iam/rbac/role/describe-json-cloud.golden"},
		{args: "iam rbac role describe CloudClusterAdmin -o yaml", fixture: "iam/rbac/role/describe-yaml-cloud.golden"},
		{args: "iam rbac role describe CloudClusterAdmin", fixture: "iam/rbac/role/describe-cloud.golden"},
		{args: "iam rbac role describe InvalidRole", fixture: "iam/rbac/role/describe-invalid-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role list -o json", fixture: "iam/rbac/role/list-json-cloud.golden"},
		{args: "iam rbac role list -o yaml", fixture: "iam/rbac/role/list-yaml-cloud.golden"},
		{args: "iam rbac role list", fixture: "iam/rbac/role/list-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRoleBinding_Cloud() {
	tests := []CLITest{
		{args: "iam rbac role-binding create --principal User:sa-12345 --role DeveloperRead --resource Topic:payroll --kafka-cluster lkc-1111aaa --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/create-service-account-developer-read.golden"},
		{args: "iam rbac role-binding create --principal User:pool-12345 --role DeveloperRead --resource Topic:payroll --kafka-cluster lkc-1111aaa --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/create-identity-pool-developer-read.golden"},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin --current-environment --cloud-cluster lkc-1111aaa"},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa"},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin", fixture: "iam/rbac/role-binding/missing-cloud-cluster-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role EnvironmentAdmin", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa --force"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --environment a-595 --cloud-cluster lkc-1111aaa", input: "y\n", fixture: "iam/rbac/role-binding/delete-prompt.golden"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --current-environment --cloud-cluster lkc-1111aaa --force"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --force", fixture: "iam/rbac/role-binding/missing-cloud-cluster-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa --force", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role EnvironmentAdmin --force", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/delete-missing-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa@confluent.io --role CloudClusterAdmin --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/create-with-email-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRoleBindingList_Cloud() {
	tests := []CLITest{
		{args: "iam rbac role-binding list", fixture: "iam/rbac/role-binding/list-no-principal-nor-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/list-no-principal-nor-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-11aaa", fixture: "iam/rbac/role-binding/list-user-1-cloud.golden"},
		{args: "iam rbac role-binding list --current-environment --cloud-cluster lkc-1111aaa --principal User:u-11aaa", fixture: "iam/rbac/role-binding/list-user-1-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-22bbb", fixture: "iam/rbac/role-binding/list-user-2-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-33ccc", fixture: "iam/rbac/role-binding/list-user-3-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-44ddd", fixture: "iam/rbac/role-binding/list-user-4-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-55eee", fixture: "iam/rbac/role-binding/list-user-5-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --principal User:u-66fff", fixture: "iam/rbac/role-binding/list-user-6-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --principal User:u-77ggg", fixture: "iam/rbac/role-binding/list-user-7-cloud.golden"},
		{args: "iam rbac role-binding list --role OrganizationAdmin", fixture: "iam/rbac/role-binding/list-user-orgadmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --role EnvironmentAdmin", fixture: "iam/rbac/role-binding/list-user-envadmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin", fixture: "iam/rbac/role-binding/list-user-clusteradmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml", fixture: "iam/rbac/role-binding/list-user-clusteradmin-yaml-cloud.golden"},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json", fixture: "iam/rbac/role-binding/list-user-clusteradmin-json-cloud.golden"},
		{args: "iam rbac role-binding list --principal User:u-41dxz3 --cluster pantsCluster", fixture: "iam/rbac/role-binding/list-failure-help-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --role InvalidOrgAdmin", fixture: "iam/rbac/role-binding/list-invalid-role-error-type-1-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment a-595 --cloud-cluster lkc-1111aaa --role InvalidMetricsViewer", fixture: "iam/rbac/role-binding/list-invalid-role-error-type-2-cloud.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRoleBinding_OnPrem() {
	tests := []CLITest{
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster", fixture: "iam/rbac/role-binding/create-cluster-name-onprem.golden"},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster kafka-GUID", fixture: "iam/rbac/role-binding/create-cluster-id-onprem.golden"},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster kafka-GUID --cluster-name theMdsConnectCluster", fixture: "iam/rbac/role-binding/name-and-id-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksqlname --cluster-name theMdsConnectCluster", fixture: "iam/rbac/role-binding/name-and-id-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs", fixture: "iam/rbac/role-binding/missing-name-or-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksql-name", fixture: "iam/rbac/role-binding/missing-kafka-cluster-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksqlName --connect-cluster connectID --kafka-cluster kafka-GUID", fixture: "iam/rbac/role-binding/multiple-non-kafka-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster --force", fixture: "iam/rbac/role-binding/delete-cluster-name-onprem.golden"},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --cluster-name theMdsConnectCluster", input: "y\n", fixture: "iam/rbac/role-binding/delete-cluster-name-onprem-prompt.golden"},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster kafka-GUID --force", fixture: "iam/rbac/role-binding/delete-cluster-id-onprem.golden"},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --kafka-cluster kafka-GUID --cluster-name theMdsConnectCluster --force", fixture: "iam/rbac/role-binding/name-and-id-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksqlname --cluster-name theMdsConnectCluster --force", fixture: "iam/rbac/role-binding/name-and-id-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --force", fixture: "iam/rbac/role-binding/missing-name-or-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksql-name --force", fixture: "iam/rbac/role-binding/missing-kafka-cluster-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource Topic:connect-configs --ksql-cluster ksqlName --connect-cluster connectID --kafka-cluster kafka-GUID --force", fixture: "iam/rbac/role-binding/multiple-non-kafka-id-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:bob@Kafka --role DeveloperRead --resource Topic:connect-configs --kafka-cluster kafka-GUID", fixture: "iam/rbac/role-binding/create-cluster-id-at-onprem.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRoleBindingList_OnPrem() {
	tests := []CLITest{
		{args: "iam rbac role-binding list --kafka-cluster CID", fixture: "iam/rbac/role-binding/list-no-principal-nor-role-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal frodo", fixture: "iam/rbac/role-binding/list-principal-format-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo", fixture: "iam/rbac/role-binding/list-user-onprem.golden"},
		{args: "iam rbac role-binding list --cluster-name kafka --principal User:frodo", fixture: "iam/rbac/role-binding/list-user-onprem.golden"},
		{args: "iam rbac role-binding list --cluster-name kafka --kafka-cluster CID --principal User:frodo", fixture: "iam/rbac/role-binding/name-and-id-error-onprem.golden", exitCode: 1},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role DeveloperRead", fixture: "iam/rbac/role-binding/list-user-and-role-with-multiple-resources-from-one-group-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role DeveloperRead -o json", fixture: "iam/rbac/role-binding/list-user-and-role-with-multiple-resources-from-one-group-json-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role DeveloperRead -o yaml", fixture: "iam/rbac/role-binding/list-user-and-role-with-multiple-resources-from-one-group-yaml-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role DeveloperWrite", fixture: "iam/rbac/role-binding/list-user-and-role-with-resources-from-multiple-groups-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role SecurityAdmin", fixture: "iam/rbac/role-binding/list-user-and-role-with-cluster-resource-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role SystemAdmin", fixture: "iam/rbac/role-binding/list-user-and-role-with-no-matches-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role SystemAdmin -o json", fixture: "iam/rbac/role-binding/list-user-and-role-with-no-matches-json-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal User:frodo --role SystemAdmin -o yaml", fixture: "iam/rbac/role-binding/list-user-and-role-with-no-matches-yaml-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal Group:hobbits --role DeveloperRead", fixture: "iam/rbac/role-binding/list-group-and-role-with-multiple-resources-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal Group:hobbits --role DeveloperWrite", fixture: "iam/rbac/role-binding/list-group-and-role-with-one-resource-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --principal Group:hobbits --role SecurityAdmin", fixture: "iam/rbac/role-binding/list-group-and-role-with-no-matches-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperRead", fixture: "iam/rbac/role-binding/list-role-with-multiple-bindings-to-one-group-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperRead -o json", fixture: "iam/rbac/role-binding/list-role-with-multiple-bindings-to-one-group-json-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperRead -o yaml", fixture: "iam/rbac/role-binding/list-role-with-multiple-bindings-to-one-group-yaml-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperWrite", fixture: "iam/rbac/role-binding/list-role-with-bindings-to-multiple-groups-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role SecurityAdmin", fixture: "iam/rbac/role-binding/list-role-on-cluster-bound-to-user-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role SystemAdmin", fixture: "iam/rbac/role-binding/list-role-with-no-matches-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperRead --resource Topic:food", fixture: "iam/rbac/role-binding/list-role-and-resource-with-exact-match-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperRead --resource Topic:shire-parties", fixture: "iam/rbac/role-binding/list-role-and-resource-with-no-match-onprem.golden"},
		{args: "iam rbac role-binding list --kafka-cluster CID --role DeveloperWrite --resource Topic:shire-parties", fixture: "iam/rbac/role-binding/list-role-and-resource-with-prefix-match-onprem.golden"},
		{args: "iam rbac role-binding list --principal User:u-41dxz3 --cluster pantsCluster", fixture: "iam/rbac/role-binding/list-failure-help-onprem.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamServiceAccount() {
	tests := []CLITest{
		{args: "iam service-account create human-service --description human-output", fixture: "iam/service-account/create.golden"},
		{args: "iam service-account create json-service --description json-output -o json", fixture: "iam/service-account/create-json.golden"},
		{args: "iam service-account create yaml-service --description yaml-output -o yaml", fixture: "iam/service-account/create-yaml.golden"},
		{args: "iam service-account delete sa-12345 --force", fixture: "iam/service-account/delete.golden"},
		{args: "iam service-account delete sa-12345", input: "service_account\n", fixture: "iam/service-account/delete-prompt.golden"},
		{args: "iam service-account list -o json", fixture: "iam/service-account/list-json.golden"},
		{args: "iam service-account list -o yaml", fixture: "iam/service-account/list-yaml.golden"},
		{args: "iam service-account list", fixture: "iam/service-account/list.golden"},
		{args: "iam service-account describe sa-12345 -o json", fixture: "iam/service-account/describe-json.golden"},
		{args: "iam service-account describe sa-12345 -o yaml", fixture: "iam/service-account/describe-yaml.golden"},
		{args: "iam service-account describe sa-12345", fixture: "iam/service-account/describe.golden"},
		{args: "iam service-account describe sa-6789", fixture: "iam/service-account/service-account-not-found.golden", exitCode: 1},
		{args: "iam service-account update sa-12345 --description new-description", fixture: "iam/service-account/update.golden"},
		{args: "iam service-account update sa-12345 --description new-description-2", fixture: "iam/service-account/update-2.golden"},
		{args: "iam service-account delete sa-12345 --force", fixture: "iam/service-account/delete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserList() {
	tests := []CLITest{
		{args: "iam user list", fixture: "iam/user/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserDescribe() {
	tests := []CLITest{
		{args: "iam user describe u-0", fixture: "iam/user/resource-not-found.golden", exitCode: 1},
		{args: "iam user describe u-17", fixture: "iam/user/describe.golden"},
		{args: "iam user describe 0", fixture: "iam/user/bad-resource-id.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserDelete() {
	tests := []CLITest{
		{args: "iam user delete u-2 --force", fixture: "iam/user/delete.golden"},
		{args: "iam user delete u-2", input: "Bono\n", fixture: "iam/user/delete-prompt.golden"},
		{args: "iam user delete 0 --force", fixture: "iam/user/bad-resource-id.golden", exitCode: 1},
		{args: "iam user delete u-1 --force", fixture: "iam/user/delete-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserUpdate() {
	tests := []CLITest{
		{args: "iam user update u-11aaa --full-name Test", fixture: "iam/user/update.golden"},
		{args: "iam user update 0 --full-name Test", fixture: "iam/user/bad-resource-id.golden", exitCode: 1},
		{args: "iam user update u-1 --full-name Test", fixture: "iam/user/update-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserInvitationCreate() {
	tests := []CLITest{
		{args: "iam user invitation create miles@confluent.io", fixture: "iam/user/invite.golden"},
		{args: "iam user invitation create bad-email.com", exitCode: 1, fixture: "iam/user/bad-email.golden"},
		{args: "iam user invitation create user@exists.com", exitCode: 1, fixture: "iam/user/invite-user-already-active.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserInvitationList() {
	tests := []CLITest{
		{args: "iam user invitation list", fixture: "iam/user/invitation_list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamProviderCreate() {
	tests := []CLITest{
		{args: "iam provider create Okta --description new-description --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com", fixture: "iam/identity-provider/create.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamProviderDelete() {
	tests := []CLITest{
		{args: "iam provider delete op-12345 --force", fixture: "iam/identity-provider/delete.golden"},
		{args: "iam provider delete op-12345", input: "identity_provider\n", fixture: "iam/identity-provider/delete-prompt.golden"},
		{args: "iam provider delete op-1 --force", fixture: "iam/identity-provider/delete-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamProviderDescribe() {
	tests := []CLITest{
		{args: "iam provider describe op-12345", fixture: "iam/identity-provider/describe.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamProviderUpdate() {
	tests := []CLITest{
		{args: "iam provider update op-12345 --name new-name --description new-description", fixture: "iam/identity-provider/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamProviderList() {
	tests := []CLITest{
		{args: "iam provider list", fixture: "iam/identity-provider/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPoolCreate() {
	tests := []CLITest{
		{args: `iam pool create testPool --provider op-12345 --description new-description --identity-claim sub --filter "claims.iss=https://company.provider.com"`, fixture: "iam/identity-pool/create.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPoolDelete() {
	tests := []CLITest{
		{args: "iam pool delete pool-55555 --provider op-12345 --force", fixture: "iam/identity-pool/delete.golden"},
		{args: "iam pool delete pool-55555 --provider op-12345", input: "identity_pool\n", fixture: "iam/identity-pool/delete-prompt.golden"},
		{args: "iam pool delete pool-1 --provider op-12345 --force", fixture: "iam/identity-pool/delete-dne.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPoolDescribe() {
	tests := []CLITest{
		{args: "iam pool describe pool-12345 --provider op-12345", fixture: "iam/identity-pool/describe.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPoolUpdate() {
	tests := []CLITest{
		{args: `iam pool update pool-12345 --provider op-12345 --name newer-name --description more-descriptive --identity-claim new-sub --filter "claims.iss=https://new-company.new-provider.com"`, fixture: "iam/identity-pool/update.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPoolList() {
	tests := []CLITest{
		{args: "iam pool list --provider op-12345", fixture: "iam/identity-pool/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamAutocomplete() {
	tests := []CLITest{
		{args: `__complete iam pool describe --provider op-12345 ""`, fixture: "iam/identity-pool/describe-autocomplete.golden"},
		{args: `__complete iam provider describe ""`, fixture: "iam/identity-provider/describe-autocomplete.golden"},
		{args: `__complete iam service-account describe ""`, fixture: "iam/service-account/describe-autocomplete.golden"},
		{args: `__complete iam user describe ""`, fixture: "iam/user/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
