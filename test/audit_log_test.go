package test

import "fmt"

func (s *CLITestSuite) TestAuditLogDescribe() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe.golden"})
}

func (s *CLITestSuite) TestAuditLogRoute() {
	tests := []CLITest{
		{args: "audit-log route list --resource crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/connect=qa-test", fixture: "audit-log/route/list.golden"},
		{args: "audit-log route lookup crn://mds1.example.com/kafka=abcde_FGHIJKL-01234567/topic=qa-test", fixture: "audit-log/route/lookup.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAuditLogConfig() {
	tests := []CLITest{
		{args: "audit-log config describe", fixture: "audit-log/config/describe.golden"},
		{args: "audit-log config update --file test/fixtures/input/audit-log/config-spec-update.json", fixture: "audit-log/config/update.golden"},
		{args: "audit-log config update --force --file test/fixtures/input/audit-log/config-spec-update.json", fixture: "audit-log/config/update-force.golden"},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAuditLogConfigMigrate() {
	migration1 := getInputFixturePath("audit-log", "config-migration-server1.golden")
	migration2 := getInputFixturePath("audit-log", "config-migration-server2.golden")

	malformed := getInputFixturePath("audit-log", "malformed-migration.golden")
	nullFields := getInputFixturePath("audit-log", "null-fields-migration.golden")

	tests := []CLITest{
		{
			args: fmt.Sprintf("audit-log config migrate --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2 --bootstrap-servers new_bootstrap_1 --authority NEW.CRN.AUTHORITY.COM", migration1, migration2),
			fixture: "audit-log/migrate/result-with-warnings.golden",
		},
		{
			args: fmt.Sprintf("audit-log config migrate --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2", malformed, migration2),
			contains: "Ignoring property file",
		},
		{
			args:    fmt.Sprintf("audit-log config migrate --combine cluster123=%s,clusterABC=%s", nullFields, nullFields),
			fixture: "audit-log/migrate/empty-result.golden",
		},
	}

	for _, test := range tests {
		test.login = "onprem"
		s.runIntegrationTest(test)
	}
}

func (s *CLITestSuite) TestAuditLogDescribe_Disabled() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe-fail.golden", disableAuditLog: true, exitCode: 1})
}
