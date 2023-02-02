package test

func (s *CLITestSuite) TestBYOK() {
	tests := []CLITest{
		// Only log in at the beginning so active env is not reset
		// tt.workflow=true so login is not reset
		// list tests
		{args: "byok list", fixture: "byok/list_1.golden", login: "cloud"},
		{args: "byok list --state IN_USE", fixture: "byok/list_2.golden"},
		{args: "byok list --provider AWS", fixture: "byok/list_3.golden"},
		{args: "byok list --state IN_USE --provider Azure", fixture: "byok/list_4.golden"},
		// register tests
		{args: "byok register arn:aws:kms:us-west-2:037803949979:key/0e2609e3-a0bf-4f39-aedf-8b1f63b16d81", fixture: "byok/register_1.golden"},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant-id 00000000-0000-0000-0000-000000000000 --key-vault-id /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault", fixture: "byok/register_2.golden"},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant-id 00000000-0000-0000-0000-000000000000", fixture: "byok/register_3.golden", wantErrCode: 1},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000", fixture: "byok/register_4.golden", wantErrCode: 1},
		// unregister tests
		{args: "byok unregister cck-001", fixture: "byok/unregister_1.golden"},
		{args: "byok unregister cck-404", fixture: "byok/unregister_2.golden", wantErrCode: 1},
	}

	resetConfiguration(s.T(), false)

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestBYOKDescribe() {
	tests := []CLITest{
		{args: "byok describe cck-001", fixture: "byok/describe-aws.golden"},
		{args: "byok describe cck-001 -o json", fixture: "byok/describe-aws-json.golden"},
		{args: "byok describe cck-003", fixture: "byok/describe-azure.golden"},
		{args: "byok describe cck-003 -o json --show-policy-command", fixture: "byok/describe-azure-json.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
