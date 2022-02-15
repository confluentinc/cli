package test

func (s *CLITestSuite) TestQuotaLimits() {
	tests := []CLITest{
		{args: "service-quota list kafka_cluster", fixture: "quotas/1.golden", login: "default"},
		{args: "service-quota list kafka_cluster --environment env-1", fixture: "quotas/2.golden"},
		{args: "service-quota list kafka_cluster --cluster lkc-1", fixture: "quotas/3.golden"},
		{args: "service-quota list kafka_cluster --quota-code quota_a ", fixture: "quotas/4.golden"},
		{args: "service-quota list kafka_cluster --quota-code quota_a -o json", fixture: "quotas/5.golden"},
		{args: "service-quota list kafka_cluster --quota-code quota_a -o yaml", fixture: "quotas/6.golden"},
	}

	resetConfiguration(s.T())

	for _, tt := range tests {
		tt.workflow = true
		s.runCcloudTest(tt)
	}
}
