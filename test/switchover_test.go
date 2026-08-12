package test

func (s *CLITestSuite) TestSwitchover() {
	tests := []CLITest{
		// SwitchoverPair
		{args: "switchover pair create prod-kafka-dr --member west=lkc-111111 --member east=lkc-222222 --active-member west", fixture: "switchover/pair/create.golden"},
		{args: "switchover pair list", fixture: "switchover/pair/list.golden"},
		{args: "switchover pair describe sw-123456", fixture: "switchover/pair/describe.golden"},
		{args: "switchover pair update sw-123456 --name renamed-dr", fixture: "switchover/pair/update.golden"},
		{args: "switchover pair failover sw-123456 --member east --type CLEAN", fixture: "switchover/pair/failover.golden"},
		{args: "switchover pair delete sw-123456 --force", fixture: "switchover/pair/delete.golden"},
		{args: "switchover pair describe sw-000000", fixture: "switchover/pair/describe-not-found.golden", exitCode: 1},
		{args: "switchover pair describe sw-123456 --output json", fixture: "switchover/pair/describe-json.golden"},

		// SwitchoverEndpoint
		{args: "switchover endpoint create prod-endpoint --switchover-pair sw-123456 --endpoint name=west-platt,resource-id=lkc-111111,type=PRIVATE --endpoint name=east-platt,resource-id=lkc-222222,type=PRIVATE", fixture: "switchover/endpoint/create.golden"},
		{args: "switchover endpoint list", fixture: "switchover/endpoint/list.golden"},
		{args: "switchover endpoint describe se-123456", fixture: "switchover/endpoint/describe.golden"},
		{args: "switchover endpoint activate se-123456", fixture: "switchover/endpoint/activate.golden"},
		{args: "switchover endpoint delete se-123456 --force", fixture: "switchover/endpoint/delete.golden"},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
