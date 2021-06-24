package test

func (s *CLITestSuite) TestCCloudAuditLogDescribe() {
	tests := []CLITest{
		{Args: "audit-log describe", Login: "default", Fixture: "auditlog/describe.golden"},
	}

	ResetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.Workflow = true
		s.RunCcloudTest(tt)
	}
}
