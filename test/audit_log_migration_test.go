package test

import (
	"fmt"
	"path/filepath"
	"runtime"
)

func (s *CLITestSuite) TestAuditConfigMigrate() {
	migration1 := getInputFixturePath("config-migration-server1.golden", s)
	migration2 := getInputFixturePath("config-migration-server2.golden", s)

	malformed := getInputFixturePath("malformed-migration.golden", s)
	nullFields := getInputFixturePath("null-fields-migration.golden", s)

	tests := []CLITest{
		{
			args: fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2 --bootstrap-servers new_bootstrap_1 --authority NEW.CRN.AUTHORITY.COM", migration1, migration2),
			fixture: "auditlog/migration-result-with-warnings.golden",
		},
		{
			args: fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2", malformed, migration2),
			contains: "Ignoring property file",
		},
		{
			args:    fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s", nullFields, nullFields),
			fixture: "auditlog/empty-migration-result.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}

func getInputFixturePath(file string, suite *CLITestSuite) string {
	_, callerFileName, _, ok := runtime.Caller(0)
	if !ok {
		suite.Fail("problems recovering caller information")
	}
	return filepath.Join(filepath.Dir(callerFileName), "fixtures", "input", "auditlog", file)
}
