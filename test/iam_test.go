package test

import "fmt"

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
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin --environment env-596 --cloud-cluster lkc-1111aaa"},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin", fixture: "iam/rbac/role-binding/missing-cloud-cluster-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa --role EnvironmentAdmin", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --environment env-596 --cloud-cluster lkc-1111aaa --force"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --environment env-596 --cloud-cluster lkc-1111aaa", input: "y\n", fixture: "iam/rbac/role-binding/delete-prompt.golden"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --current-environment --cloud-cluster lkc-1111aaa --force"},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --force", fixture: "iam/rbac/role-binding/missing-cloud-cluster-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role CloudClusterAdmin --cloud-cluster lkc-1111aaa --force", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --role EnvironmentAdmin --force", fixture: "iam/rbac/role-binding/missing-environment-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding delete --principal User:u-11aaa --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/delete-missing-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding create --principal User:u-11aaa@confluent.io --role CloudClusterAdmin --current-environment --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/create-with-email-cloud.golden"},
		{args: "iam rbac role-binding create --principal User:u-77ggg --role FlinkDeveloper --environment env-596 --flink-region aws.us-east-1 --resource ComputePool:lfcp-1111aaa", fixture: "iam/rbac/role-binding/create-flink-developer-cloud.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamRbacRoleBindingList_Cloud() {
	tests := []CLITest{
		{args: "iam rbac role-binding list", fixture: "iam/rbac/role-binding/list-no-principal-nor-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa", fixture: "iam/rbac/role-binding/list-no-principal-nor-role-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-11aaa", fixture: "iam/rbac/role-binding/list-user-1-cloud.golden"},
		{args: "iam rbac role-binding list --current-environment --cloud-cluster lkc-1111aaa --principal User:u-11aaa", fixture: "iam/rbac/role-binding/list-user-1-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-22bbb", fixture: "iam/rbac/role-binding/list-user-2-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-33ccc", fixture: "iam/rbac/role-binding/list-user-3-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-44ddd", fixture: "iam/rbac/role-binding/list-user-4-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-55eee", fixture: "iam/rbac/role-binding/list-user-5-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --principal User:u-66fff", fixture: "iam/rbac/role-binding/list-user-6-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --principal User:u-77ggg", fixture: "iam/rbac/role-binding/list-user-7-cloud.golden"},
		{args: "iam rbac role-binding list --role OrganizationAdmin", fixture: "iam/rbac/role-binding/list-user-orgadmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --role EnvironmentAdmin", fixture: "iam/rbac/role-binding/list-user-envadmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin", fixture: "iam/rbac/role-binding/list-user-clusteradmin-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o yaml", fixture: "iam/rbac/role-binding/list-user-clusteradmin-yaml-cloud.golden"},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --role CloudClusterAdmin -o json", fixture: "iam/rbac/role-binding/list-user-clusteradmin-json-cloud.golden"},
		{args: "iam rbac role-binding list --principal User:u-41dxz3 --cluster pantsCluster", fixture: "iam/rbac/role-binding/list-failure-help-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --role InvalidOrgAdmin", fixture: "iam/rbac/role-binding/list-invalid-role-error-type-1-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --environment env-596 --cloud-cluster lkc-1111aaa --role InvalidMetricsViewer", fixture: "iam/rbac/role-binding/list-invalid-role-error-type-2-cloud.golden", exitCode: 1},
		{args: "iam rbac role-binding list --role FlinkDeveloper --environment env-596 --flink-region aws.us-east-1 --resource ComputePool:lfcp-1111aaa", fixture: "iam/rbac/role-binding/list-flink-developer-cloud.golden"},
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
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource FlinkEnvironment:testEnvironment --cmf testCmf", fixture: "iam/rbac/role-binding/create-cmf-resource-onprem.golden"},
		{args: "iam rbac role-binding create --principal User:bob --role DeveloperRead --resource FlinkApplication:testApplication --cmf testCmf --flink-environment testFLINKEnv ", fixture: "iam/rbac/role-binding/create-flink-environment-resource-onprem.golden"},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource FlinkEnvironment:testEnvironment --cmf testCmf --force", fixture: "iam/rbac/role-binding/delete-cmf-resource-onprem.golden"},
		{args: "iam rbac role-binding delete --principal User:bob --role DeveloperRead --resource FlinkApplication:testApplication --cmf testCmf --flink-environment testFLINKEnv --force", fixture: "iam/rbac/role-binding/delete-flink-environment-resource-onprem.golden"},
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
		{args: "iam service-account create human-service --description human-output --resource-owner u-123", fixture: "iam/service-account/create.golden"},
		{args: "iam service-account create json-service --description json-output -o json", fixture: "iam/service-account/create-json.golden"},
		{args: "iam service-account create yaml-service --description yaml-output -o yaml", fixture: "iam/service-account/create-yaml.golden"},
		{args: "iam service-account delete sa-12345 --force", fixture: "iam/service-account/delete.golden"},
		{args: "iam service-account delete sa-12345 sa-67890", fixture: "iam/service-account/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam service-account delete sa-12345 sa-54321", input: "n\n", fixture: "iam/service-account/delete-multiple-refuse.golden"},
		{args: "iam service-account delete sa-12345 sa-54321", input: "y\n", fixture: "iam/service-account/delete-multiple-success.golden"},
		{args: "iam service-account delete sa-12345", input: "y\n", fixture: "iam/service-account/delete-prompt.golden"},
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

	tests = []CLITest{
		{args: "iam user describe", fixture: "iam/user/describe-onprem.golden"},
		{args: "iam user describe -o json", fixture: "iam/user/describe-onprem-json.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamUserDelete() {
	tests := []CLITest{
		{args: "iam user delete u-2 --force", fixture: "iam/user/delete.golden"},
		{args: "iam user delete u-11aaa u-1", fixture: "iam/user/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam user delete u-11aaa u-22bbb", input: "n\n", fixture: "iam/user/delete-multiple-refuse.golden"},
		{args: "iam user delete u-11aaa u-22bbb", input: "y\n", fixture: "iam/user/delete-multiple-success.golden"},
		{args: "iam user delete u-2", input: "y\n", fixture: "iam/user/delete-prompt.golden"},
		{args: "iam user delete 0 --force", fixture: "iam/user/bad-resource-id-delete.golden", exitCode: 1},
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

func (s *CLITestSuite) TestIamProvider() {
	tests := []CLITest{
		{args: "iam provider create okta --description 'new description' --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com", fixture: "iam/identity-provider/create.golden"},
		{args: "iam provider create okta-with-identity-claim --description 'new description' --jwks-uri https://company.provider.com/oauth2/v1/keys --issuer-uri https://company.provider.com --identity-claim claims.sub", fixture: "iam/identity-provider/create-with-identity-claim.golden"},
		{args: "iam provider delete op-12345 --force", fixture: "iam/identity-provider/delete.golden"},
		{args: "iam provider delete op-12345 op-54321", fixture: "iam/identity-provider/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam provider delete op-12345 op-67890", input: "n\n", fixture: "iam/identity-provider/delete-multiple-refuse.golden"},
		{args: "iam provider delete op-12345 op-67890", input: "y\n", fixture: "iam/identity-provider/delete-multiple-success.golden"},
		{args: "iam provider delete op-12345", input: "y\n", fixture: "iam/identity-provider/delete-prompt.golden"},
		{args: "iam provider delete op-1 --force", fixture: "iam/identity-provider/delete-dne.golden", exitCode: 1},
		{args: "iam provider describe op-12345", fixture: "iam/identity-provider/describe.golden"},
		{args: "iam provider describe op-67890", fixture: "iam/identity-provider/describe-with-identity-claim.golden"},
		{args: "iam provider update op-12345 --name updated-name --description 'updated description'", fixture: "iam/identity-provider/update.golden"},
		{args: "iam provider update op-67890 --identity-claim claims.sub.updated", fixture: "iam/identity-provider/update-with-identity-claim.golden"},
		{args: "iam provider list", fixture: "iam/identity-provider/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamPool() {
	tests := []CLITest{
		{args: `iam pool create test-pool --provider op-12345 --description "new description" --identity-claim sub --filter "claims.iss=https://company.provider.com"`, fixture: "iam/pool/create.golden"},
		{args: "iam pool delete pool-55555 --provider op-12345 --force", fixture: "iam/pool/delete.golden"},
		{args: "iam pool delete pool-55555 pool-44444 --provider op-12345", fixture: "iam/pool/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam pool delete pool-55555 pool-12345 --provider op-12345", input: "n\n", fixture: "iam/pool/delete-multiple-refuse.golden"},
		{args: "iam pool delete pool-55555 pool-12345 --provider op-12345", input: "y\n", fixture: "iam/pool/delete-multiple-success.golden"},
		{args: "iam pool delete pool-55555 --provider op-12345", input: "y\n", fixture: "iam/pool/delete-prompt.golden"},
		{args: "iam pool delete pool-1 --provider op-12345 --force", fixture: "iam/pool/delete-dne.golden", exitCode: 1},
		{args: "iam pool describe pool-12345 --provider op-12345", fixture: "iam/pool/describe.golden"},
		{args: `iam pool update pool-12345 --provider op-12345 --name "updated name" --description "updated description" --identity-claim new-sub --filter "claims.iss=https://new-company.new-provider.com"`, fixture: "iam/pool/update.golden"},
		{args: "iam pool update pool-12345 --provider op-12345", fixture: "iam/pool/no-op-update.golden", exitCode: 1},
		{args: "iam pool list --provider op-12345", fixture: "iam/pool/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamCertificateAuthority() {
	tests := []CLITest{
		{args: `iam certificate-authority create my-ca --description "my certificate authority" --certificate-chain ABC123 --certificate-chain-filename certificate.pem`, fixture: "iam/certificate-authority/create.golden"},
		{args: `iam certificate-authority create my-ca --description "my certificate authority" --certificate-chain ABC123 --certificate-chain-filename certificate.pem --crl-chain DEF456`, fixture: "iam/certificate-authority/create-url-chain.golden"},
		{args: "iam certificate-authority delete op-12345 --force", fixture: "iam/certificate-authority/delete.golden"},
		{args: "iam certificate-authority delete op-12345 op-67890", fixture: "iam/certificate-authority/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam certificate-authority delete op-12345 op-54321", input: "y\n", fixture: "iam/certificate-authority/delete-multiple-success.golden"},
		{args: "iam certificate-authority describe op-12345", fixture: "iam/certificate-authority/describe.golden"},
		{args: "iam certificate-authority describe op-12345 -o json", fixture: "iam/certificate-authority/describe-json.golden"},
		{args: `iam certificate-authority update op-12345 --name "new name" --description "new description" --certificate-chain ABC123 --certificate-chain-filename certificate-2.pem`, fixture: "iam/certificate-authority/update.golden"},
		{args: `iam certificate-authority update op-12345 --name "new name" --description "new description" --certificate-chain ABC123 --certificate-chain-filename certificate-2.pem --crl-url example.url`, fixture: "iam/certificate-authority/update-crl-url.golden"},
		{args: `iam certificate-authority update op-12345 --name "new name" --description "new description" --certificate-chain-filename certificate-2.pem`, fixture: "iam/certificate-authority/update-fail.golden", exitCode: 1},
		{args: "iam certificate-authority list", fixture: "iam/certificate-authority/list.golden"},
		{args: "iam certificate-authority list -o json", fixture: "iam/certificate-authority/list-json.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamCertificatePool() {
	tests := []CLITest{
		{args: `iam certificate-pool create pool-xyz --provider pool-1 --description "new description" --external-identifier "identity"`, fixture: "iam/certificate-pool/create.golden"},
		{args: "iam certificate-pool delete pool-55555 --provider pool-1 --force", fixture: "iam/certificate-pool/delete.golden"},
		{args: "iam certificate-pool delete pool-55555 pool-44444 --provider pool-1", fixture: "iam/certificate-pool/delete-multiple-fail.golden", exitCode: 1},
		{args: "iam certificate-pool delete pool-55555 pool-12345 --provider pool-1", input: "y\n", fixture: "iam/certificate-pool/delete-multiple-success.golden"},
		{args: "iam certificate-pool describe pool-12345 --provider pool-1", fixture: "iam/certificate-pool/describe.golden"},
		{args: `iam certificate-pool update pool-12345 --provider pool-1 --name "updated name" --description "updated description" --external-identifier "identity2" --filter false`, fixture: "iam/certificate-pool/update.golden"},
		{args: "iam certificate-pool list --provider pool-1", fixture: "iam/certificate-pool/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamGroupMapping() {
	tests := []CLITest{
		{args: `iam group-mapping create group_mapping --description new-group-description --filter '"engineering" in claims.group || "marketing" in claims.group'`, fixture: "iam/group-mapping/create.golden"},
		{args: `iam group-mapping create group-mapping-rate-limit`, fixture: "iam/group-mapping/create-402-error.golden", exitCode: 1},
		{args: "iam group-mapping delete group-abc --force", fixture: "iam/group-mapping/delete.golden"},
		{args: "iam group-mapping delete group-abc", input: "y\n", fixture: "iam/group-mapping/delete-prompt.golden"},
		{args: "iam group-mapping delete group-abc group-def", input: "n\n", fixture: "iam/group-mapping/delete-multiple-refuse.golden"},
		{args: "iam group-mapping delete group-abc group-def", input: "y\n", fixture: "iam/group-mapping/delete-multiple-success.golden"},
		{args: "iam group-mapping delete group-dne --force", fixture: "iam/group-mapping/delete-dne.golden", exitCode: 1},
		{args: "iam group-mapping describe group-abc", fixture: "iam/group-mapping/describe.golden"},
		{args: `iam group-mapping update group-abc --name updated-group-mapping --description "updated description" --filter claims.principal.startsWith("user")`, fixture: "iam/group-mapping/update.golden"},
		{args: "iam group-mapping list", fixture: "iam/group-mapping/list.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIam_Autocomplete() {
	tests := []CLITest{
		{args: `__complete iam pool describe --provider op-12345 ""`, fixture: "iam/pool/describe-autocomplete.golden"},
		{args: `__complete iam provider describe ""`, fixture: "iam/identity-provider/describe-autocomplete.golden"},
		{args: `__complete iam service-account describe ""`, fixture: "iam/service-account/describe-autocomplete.golden"},
		{args: `__complete iam user describe ""`, fixture: "iam/user/describe-autocomplete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamIpGroup() {
	tests := []CLITest{
		{args: "iam ip-group create demo-ip-group --cidr-blocks 168.150.200.0/24,147.150.200.0/24", fixture: "iam/ip-group/create.golden"},
		{args: "iam ip-group list", fixture: "iam/ip-group/list.golden"},
		{args: "iam ip-group describe ipg-wjnde", fixture: "iam/ip-group/describe.golden"},
		{args: "iam ip-group delete ipg-wjnde", fixture: "iam/ip-group/delete.golden"},
		{args: "iam ip-group update ipg-wjnde --name new-demo-group --add-cidr-blocks 1.2.3.4/12 --remove-cidr-blocks 168.150.200.0/24", fixture: "iam/ip-group/update.golden"},
		{args: "iam ip-group update ipg-wjnde --name new-demo-group --add-cidr-blocks 1.2.3.4/12,147.150.200.0/24 --remove-cidr-blocks 168.150.200.0/24", fixture: "iam/ip-group/update-resource-duplicate.golden"},
		{args: "iam ip-group update ipg-wjnde --name new-demo-group --add-cidr-blocks 1.2.3.4/12 --remove-cidr-blocks 1.2.3.4/12", fixture: "iam/ip-group/update-resource-add-and-remove.golden"},
		{args: "iam ip-group update ipg-wjnde --name new-demo-group --add-cidr-blocks 1.2.3.4/12 --remove-cidr-blocks 1.1.1.1/1", fixture: "iam/ip-group/update-resource-remove-not-exist.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamIpFilter() {
	tests := []CLITest{
		{args: "iam ip-filter create demo-ip-filter1 --resource-group management --ip-groups ipg-3g5jw,ipg-wjnde", fixture: "iam/ip-filter/create.golden"},
		{args: "iam ip-filter create demo-ip-filter3 --operations MANAGEMENT --ip-groups ipg-3g5jw,ipg-wjnde", fixture: "iam/ip-filter/create-operation-groups.golden"},
		{args: "iam ip-filter create demo-ip-filter4 --operations MANAGEMENT --no-public-networks", fixture: "iam/ip-filter/create-npn-group.golden"},
		{args: "iam ip-filter list", fixture: "iam/ip-filter/list.golden"},
		{args: "iam ip-filter describe ipf-34dq3", fixture: "iam/ip-filter/describe.golden"},
		{args: "iam ip-filter delete ipf-34dq3", fixture: "iam/ip-filter/delete.golden"},
		{args: "iam ip-filter update ipf-34dq3 --name new-ip-filter-demo --add-ip-groups ipg-1337a,ipg-ayd3n --remove-ip-groups ipg-12345", fixture: "iam/ip-filter/update.golden"},
		{args: "iam ip-filter update ipf-34dq3 --add-ip-groups ipg-abcde --remove-ip-groups ipg-12345", fixture: "iam/ip-filter/update-resource-duplicate.golden"},
		{args: "iam ip-filter update ipf-34dq3 --add-ip-groups ipg-azbye --remove-ip-groups ipg-azbye", fixture: "iam/ip-filter/update-resource-add-and-remove.golden"},
		{args: "iam ip-filter update ipf-34dq3 --add-ip-groups ipg-hjkil --remove-ip-groups ipg-fedbc", fixture: "iam/ip-filter/update-resource-remove-not-exist.golden"},
		{args: "iam ip-filter update ipf-34dq5 --resource-group multiple --add-operation-groups SCHEMA", fixture: "iam/ip-filter/update-add-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq4 --remove-operation-groups SCHEMA", fixture: "iam/ip-filter/update-remove-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq4 --resource-group multiple --add-operation-groups FLINK", fixture: "iam/ip-filter/update-add-flink-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq6 --remove-operation-groups SCHEMA,FLINK", fixture: "iam/ip-filter/update-remove-sr-and-flink-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq4 --resource-group multiple --add-operation-groups KAFKA_MANAGEMENT,KAFKA_DATA,KAFKA_DISCOVERY", fixture: "iam/ip-filter/update-add-kafka-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq7 --remove-operation-groups KAFKA_MANAGEMENT,KAFKA_DATA,KAFKA_DISCOVERY", fixture: "iam/ip-filter/update-remove-kafka-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq4 --resource-group multiple --add-operation-groups KSQL", fixture: "iam/ip-filter/update-add-ksql-operation-group.golden"},
		{args: "iam ip-filter update ipf-34dq8 --remove-operation-groups KSQL", fixture: "iam/ip-filter/update-remove-ksql-operation-group.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}

var mdsResourcePatterns = []struct {
	Args string
	Name string
}{
	{
		Args: "--cluster-scope",
		Name: "cluster-scope",
	},
	{
		Args: "--topic test-topic",
		Name: "topic",
	},
	{
		Args: "--topic test-topic --prefix",
		Name: "topic-prefix",
	},
	{
		Args: "--consumer-group test-group",
		Name: "consumer-group",
	},
	{
		Args: "--consumer-group test-group --prefix",
		Name: "consumer-group-prefix",
	},
	{
		Args: "--transactional-id test-transactional-id",
		Name: "transactional-id",
	},
	{
		Args: "--transactional-id test-transactional-id --prefix",
		Name: "transactional-id-prefix",
	},
}

var mdsAclEntries = []struct {
	Args string
	Name string
}{
	{
		Args: "--allow --principal User:42 --operation read",
		Name: "allow-principal-user-operation-read",
	},
	{
		Args: "--deny --principal User:42 --host testhost --operation read",
		Name: "deny-principal-host-operation-read",
	},
	{
		Args: "--allow --principal User:42 --host * --operation write",
		Name: "allow-principal-host-star-operation-write",
	},
	{
		Args: "--deny --principal User:42 --operation write",
		Name: "deny-principal-user-operation-write",
	},
	{
		Args: "--allow --principal User:42 --operation create",
		Name: "allow-principal-user-operation-create",
	},
	{
		Args: "--deny --principal User:42 --operation create",
		Name: "deny-principal-user-operation-create",
	},
	{
		Args: "--allow --principal User:42 --operation delete",
		Name: "allow-principal-user-operation-delete",
	},
	{
		Args: "--deny --principal User:42 --operation delete",
		Name: "deny-principal-user-operation-delete",
	},
	{
		Args: "--allow --principal User:42 --operation alter",
		Name: "allow-principal-user-operation-alter",
	},
	{
		Args: "--deny --principal User:42 --operation alter",
		Name: "deny-principal-user-operation-alter",
	},
	{
		Args: "--allow --principal User:42 --operation describe",
		Name: "allow-principal-user-operation-describe",
	},
	{
		Args: "--deny --principal User:42 --operation describe",
		Name: "deny-principal-user-operation-describe",
	},
	{
		Args: "--allow --principal User:42 --operation cluster-action",
		Name: "allow-principal-user-operation-cluster-action",
	},
	{
		Args: "--deny --principal User:42 --operation cluster-action",
		Name: "deny-principal-user-operation-cluster-action",
	},
	{
		Args: "--allow --principal User:42 --operation describe-configs",
		Name: "allow-principal-user-operation-describe-configs",
	},
	{
		Args: "--deny --principal User:42 --operation describe-configs",
		Name: "deny-principal-user-operation-describe-configs",
	},
	{
		Args: "--allow --principal User:42 --operation alter-configs",
		Name: "allow-principal-user-operation-alter-configs",
	},
	{
		Args: "--deny --principal User:42 --operation alter-configs",
		Name: "deny-principal-user-operation-alter-configs",
	},
	{
		Args: "--allow --principal User:42 --operation idempotent-write",
		Name: "allow-principal-user-operation-idempotent-write",
	},
	{
		Args: "--deny --principal User:42 --operation idempotent-write",
		Name: "deny-principal-user-operation-idempotent-write",
	},
}

func (s *CLITestSuite) TestIamAclList() {
	tests := []CLITest{
		{args: "iam acl list --kafka-cluster testcluster --principal User:42", fixture: "iam/acl/list-principal-user.golden"},
	}
	for _, mdsResourcePattern := range mdsResourcePatterns {
		tests = append(tests, CLITest{args: fmt.Sprintf("iam acl list --kafka-cluster testcluster %s", mdsResourcePattern.Args), fixture: fmt.Sprintf("iam/acl/list-%s.golden", mdsResourcePattern.Name)})
		tests = append(tests, CLITest{args: fmt.Sprintf("iam acl list --kafka-cluster testcluster %s --principal User:42", mdsResourcePattern.Args), fixture: fmt.Sprintf("iam/acl/list-%s-principal-user.golden", mdsResourcePattern.Name)})
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamAclCreate() {
	tests := []CLITest{
		{args: "iam acl create --kafka-cluster testcluster --allow --operation read --principal User:42 --topic resource1 --consumer-group resource2", fixture: "iam/acl/create-exactly-one-set-error.golden", exitCode: 1},
	}
	for _, mdsAclEntry := range mdsAclEntries {
		tests = append(tests, CLITest{args: fmt.Sprintf("iam acl create --kafka-cluster testcluster --cluster-scope %s", mdsAclEntry.Args), fixture: fmt.Sprintf("iam/acl/create-cluster-scope-%s.golden", mdsAclEntry.Name)})
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestIamAclDelete() {
	tests := []CLITest{
		{args: `iam acl delete --kafka-cluster testcluster --cluster-scope --principal User:abc123 --operation write --host "*" --force`, fixture: "iam/acl/delete.golden"},
		{args: `iam acl delete --kafka-cluster testcluster --principal User:def456 --operation any --host "*" --force`, fixture: "iam/acl/delet-multiple.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}
