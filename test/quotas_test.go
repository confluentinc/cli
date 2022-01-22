package test

func (s *CLITestSuite) TestQuotaLimits() {
	tests := []CLITest{
		// only login at the begginning so active env is not reset
		// tt.workflow=true so login is not reset
		{args: "quota-limits list kafka_cluster", fixture: "quotas/1.golden", login: "default"},

	}

	resetConfiguration(s.T(), "ccloud")

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}
