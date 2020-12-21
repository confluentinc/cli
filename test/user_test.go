package test

func (s *CLITestSuite) TestUserList() {
	tests := []CLITest{
		{
			args:    "admin user list",
			fixture: "admin/user-list.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserDescribe() {
	tests := []CLITest{
		{
			args:        "admin user describe u-0",
			wantErrCode: 1,
			fixture:     "admin/user-resource-not-found.golden",
		},
		{
			args:    "admin user describe u-17",
			fixture: "admin/user-describe.golden",
		},
		{
			args:        "admin user describe 0",
			wantErrCode: 1,
			fixture:     "admin/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserDelete() {
	tests := []CLITest{
		{
			args:    "admin user delete u-0",
			fixture: "admin/user-delete.golden",
		},
		{
			args:        "admin user delete 0",
			wantErrCode: 1,
			fixture:     "admin/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserInvite() {
	tests := []CLITest{
		{
			args:    "admin user invite miles@confluent.io",
			fixture: "admin/user-invite.golden",
		},
		{
			args:        "admin user invite bad-email.com",
			wantErrCode: 1,
			fixture:     "admin/user-bad-email.golden",
		},
		{
			args:        "admin user invite test@error.io",
			wantErrCode: 1,
			fixture:     "admin/user-invite-generic-error.golden",
		},
	}

	for _, test := range tests {
		test.login = "default"
		s.runCcloudTest(test)
	}
}
