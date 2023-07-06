package test

import (
	"fmt"
)

func (s *CLITestSuite) TestAuditLogDescribe() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe.golden"})
}

func (s *CLITestSuite) TestAuditConfigMigrate() {
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

	for _, tt := range tests {
		tt.login = "onprem"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestAuditLogDisabledDescribe() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe-fail.golden", disableAuditLog: true, exitCode: 1})
}
