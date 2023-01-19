package test

func (s *CLITestSuite) TestBYOK() {
	tests := []CLITest{
		// Only log in at the beginning so active env is not reset
		// tt.workflow=true so login is not reset
		{args: "byok list", fixture: "byok/1.golden", login: "cloud"},
		{args: "byok register arn:aws:kms:us-west-2:037803949979:key/0e2609e3-a0bf-4f39-aedf-8b1f63b16d81", fixture: "byok/2.golden"},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant_id 00000000-0000-0000-0000-000000000000 --key_vault_id /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault", fixture: "byok/3.golden"},
		{args: "byok list", fixture: "byok/4.golden"},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant_id 00000000-0000-0000-0000-000000000000", fixture: "byok/5.golden", wantErrCode: 1},
		{args: "byok register https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000", fixture: "byok/6.golden", wantErrCode: 1},
		{args: "byok unregister cck-004", fixture: "byok/7.golden"},
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
		{args: "byok describe cck-003 -o json", fixture: "byok/describe-azure-json.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
