package test

func (s *CLITestSuite) TestServiceAccount() {
	tests := []CLITest{
		{args: "service-account create human-service --description human-output", fixture: "service-account/service-account-create.golden"},
		{args: "service-account create json-service --description json-output -o json", fixture: "service-account/service-account-create-json.golden"},
		{args: "service-account create yaml-service --description yaml-output -o yaml", fixture: "service-account/service-account-create-yaml.golden"},
		{args: "service-account delete 12345", fixture: "service-account/service-account-delete.golden"},
		{args: "service-account list -o json", fixture: "service-account/service-account-list-json.golden"},
		{args: "service-account list -o yaml", fixture: "service-account/service-account-list-yaml.golden"},
		{args: "service-account list", fixture: "service-account/service-account-list.golden"},
		{args: "service-account update 12345 --description new-description", fixture: "service-account/service-account-update.golden"},
		{args: "service-account update sa-12345 --description new-description-2", fixture: "service-account/service-account-update-2.golden"},
		{args: "service-account delete sa-12345", fixture: "service-account/service-account-delete.golden"},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runCcloudTest(tt)
	}
}
