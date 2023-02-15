package test

import (
	"strings"

	"github.com/confluentinc/bincover"
)

func (s *CLITestSuite) TestBYOK() {
	tests := []CLITest{
		// Only log in at the beginning so active env is not reset
		// tt.workflow=true so login is not reset
		// list tests
		{args: "byok list", fixture: "byok/list_1.golden", login: "cloud"},
		{args: "byok list --state IN_USE", fixture: "byok/list_2.golden"},
		{args: "byok list --provider AWS", fixture: "byok/list_3.golden"},
		{args: "byok list --state IN_USE --provider Azure", fixture: "byok/list_4.golden"},
		// create tests
		{args: "byok create arn:aws:kms:us-west-2:037803949979:key/0e2609e3-a0bf-4f39-aedf-8b1f63b16d81", fixture: "byok/create_1.golden"},
		{args: "byok create https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant 00000000-0000-0000-0000-000000000000 --key-vault /subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/a-resourcegroups/providers/Microsoft.KeyVault/vaults/a-vault", fixture: "byok/create_2.golden"},
		{args: "byok create https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000 --tenant 00000000-0000-0000-0000-000000000000", fixture: "byok/create_3.golden", wantErrCode: 1},
		{args: "byok create https://a-vault.vault.azure.net/keys/a-key/00000000000000000000000000000000", fixture: "byok/create_4.golden", wantErrCode: 1},
		// delete tests
		{args: "byok delete cck-001", preCmdFuncs: []bincover.PreCmdFunc{stdinPipeFunc(strings.NewReader("y\n"))}, fixture: "byok/delete_1.golden"},
		{args: "byok delete cck-404", fixture: "byok/delete_2.golden", wantErrCode: 1},
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
		{args: "byok describe cck-003 -o json --policy-command", fixture: "byok/describe-azure-json.golden"},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}
