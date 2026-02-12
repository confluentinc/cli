package test

func (s *CLITestSuite) TestOrganizationSCIM() {
	tests := []CLITest{
		{args: "organization scim-token create", fixture: "organization/scim-token-create.golden"},
		{args: "organization scim-token create -o json", fixture: "organization/scim-token-create-json.golden"},
		{args: "organization scim-token create --expire-duration-mins 43200", fixture: "organization/scim-token-create-custom-expiration.golden"},
		{args: "organization scim-token list", fixture: "organization/scim-token-list.golden"},
		{args: "organization scim-token list -o json", fixture: "organization/scim-token-list-json.golden"},
		{args: "organization scim-token delete scim_token-12345 --force", fixture: "organization/scim-token-delete.golden"},
		{args: "organization scim-token delete scim_token-67890", input: "y\n", fixture: "organization/scim-token-delete-prompt.golden"},
		{args: "organization scim-token delete nonexistent --force", fixture: "organization/scim-token-delete-not-found.golden", exitCode: 1},
	}

	for _, test := range tests {
		test.login = "cloud"
		s.runIntegrationTest(test)
	}
}
