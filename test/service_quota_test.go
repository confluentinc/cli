package test

func (s *CLITestSuite) TestServiceQuotaList() {
	tests := []CLITest{
		{args: "service-quota list kafka-cluster", fixture: "service-quota/1.golden", login: "cloud"},
		{args: "service-quota list kafka-cluster --environment env-1", fixture: "service-quota/2.golden"},
		{args: "service-quota list kafka-cluster --cluster lkc-1", fixture: "service-quota/3.golden"},
		{args: "service-quota list kafka-cluster --quota-code quota_a", fixture: "service-quota/4.golden"},
		{args: "service-quota list kafka-cluster --quota-code quota_a -o json", fixture: "service-quota/5.golden"},
		{args: "service-quota list kafka-cluster --quota-code quota_a -o yaml", fixture: "service-quota/6.golden"},
	}

	resetConfiguration(s.T(), false)

	for _, test := range tests {
		test.workflow = true
		s.runIntegrationTest(test)
	}
}
