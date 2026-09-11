package test

func (s *CLITestSuite) TestSwitchover() {
	tests := []CLITest{
		// SwitchoverPair
		{args: `switchover pair create prod-kafka-dr --member name=west,crn=crn://confluent.cloud/organization=org-123/environment=env-123456/cloud-cluster=lkc-111111 --member name=east,crn=crn://confluent.cloud/organization=org-123/environment=env-234567/cloud-cluster=lkc-222222 --active-member west`, fixture: "switchover/pair/create.golden"},
		{args: "switchover pair list", fixture: "switchover/pair/list.golden"},
		{args: "switchover pair describe sw-123456", fixture: "switchover/pair/describe.golden"},
		{args: "switchover pair update sw-123456 --display-name renamed-dr", fixture: "switchover/pair/update.golden"},
		{args: "switchover pair trigger-switch sw-123456 --active-member east --failover-type UNPLANNED --force", fixture: "switchover/pair/trigger-switch.golden"},
		{args: "switchover pair delete sw-123456 --force", fixture: "switchover/pair/delete.golden"},
		{args: "switchover pair describe sw-000000", fixture: "switchover/pair/describe-not-found.golden", exitCode: 1},
		{args: "switchover pair describe sw-123456 --output json", fixture: "switchover/pair/describe-json.golden"},

		// SwitchoverEndpoint
		{args: `switchover endpoint create prod-endpoint --switchover-pair sw-123456 --endpoint name=west-platt,type=private,network=n-111111 --endpoint name=east-platt,type=private,network=n-222222`, fixture: "switchover/endpoint/create.golden"},
		{args: "switchover endpoint list", fixture: "switchover/endpoint/list.golden"},
		{args: "switchover endpoint describe se-123456", fixture: "switchover/endpoint/describe.golden"},
		{args: "switchover endpoint update se-123456 --display-name renamed-endpoint", fixture: "switchover/endpoint/update.golden"},
		{args: "switchover endpoint delete se-123456 --force", fixture: "switchover/endpoint/delete.golden"},
		{args: "switchover endpoint describe se-000000", fixture: "switchover/endpoint/describe-not-found.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
