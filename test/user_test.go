package test

func (s *CLITestSuite) TestUserList() {
	tests := []CLITest{
		{
			Args:    "admin user list",
			Fixture: "admin/user-list.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserDescribe() {
	tests := []CLITest{
		{
			Args:        "admin user describe u-0",
			WantErrCode: 1,
			Fixture:     "admin/user-resource-not-found.golden",
		},
		{
			Args:    "admin user describe u-17",
			Fixture: "admin/user-describe.golden",
		},
		{
			Args:        "admin user describe 0",
			WantErrCode: 1,
			Fixture:     "admin/user-bad-resource-id.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserDelete() {
	tests := []CLITest{
		{
			Args:    "admin user delete u-0",
			Fixture: "admin/user-delete.golden",
		},
		{
			Args:        "admin user delete 0",
			WantErrCode: 1,
			Fixture:     "admin/user-bad-resource-id.golden",
		},
		{
			Args:        "admin user delete u-1",
			WantErrCode: 1,
			Fixture:     "admin/user-delete-dne.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}

func (s *CLITestSuite) TestUserInvite() {
	tests := []CLITest{
		{
			Args:    "admin user invite miles@confluent.io",
			Fixture: "admin/user-invite.golden",
		},
		{
			Args:        "admin user invite bad-email.com",
			WantErrCode: 1,
			Fixture:     "admin/user-bad-email.golden",
		},
		{
			Args:        "admin user invite user@exists.com",
			WantErrCode: 1,
			Fixture:     "admin/user-invite-user-already-active.golden",
		},
	}

	for _, test := range tests {
		test.Login = "default"
		s.RunCcloudTest(test)
	}
}
