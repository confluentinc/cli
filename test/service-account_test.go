package test

func (s *CLITestSuite) TestServiceAccount() {
	tests := []CLITest{
		{Args: "service-account create human-service --description human-output", Fixture: "service-account/service-account-create.golden"},
		{Args: "service-account create json-service --description json-output -o json", Fixture: "service-account/service-account-create-json.golden"},
		{Args: "service-account create yaml-service --description yaml-output -o yaml", Fixture: "service-account/service-account-create-yaml.golden"},
		{Args: "service-account delete 12345", Fixture: "service-account/service-account-delete.golden"},
		{Args: "service-account list -o json", Fixture: "service-account/service-account-list-json.golden"},
		{Args: "service-account list -o yaml", Fixture: "service-account/service-account-list-yaml.golden"},
		{Args: "service-account list", Fixture: "service-account/service-account-list.golden"},
		{Args: "service-account update 12345 --description new-description", Fixture: "service-account/service-account-update.golden"},
		{Args: "service-account update sa-12345 --description new-description-2", Fixture: "service-account/service-account-update-2.golden"},
		{Args: "service-account delete sa-12345", Fixture: "service-account/service-account-delete.golden"},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunCcloudTest(tt)
	}
}
