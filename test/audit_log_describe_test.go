package test

func (s *CLITestSuite) TestCCloudAuditLogDescribe() {
	tests := []CLITest{
		{args: "audit-log describe", login: "default", fixture: "auditlog/describe.golden"},
	}

	resetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}
