package test

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"regexp"
	"runtime"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

func (s *CLITestSuite) TestAuditLogConfig() {
	tests := []CLITest{
		{
			name:    "confluent audit-log config describe --help",
			args:    "audit-log config describe --help",
			fixture: "auditlog/confluent-audit-log-config-describe-help.golden",
		},
		{
			name:    "confluent audit-log config edit --help",
			args:    "audit-log config edit --help",
			fixture: "auditlog/confluent-audit-log-config-edit-help.golden",
		},
		{
			name:    "confluent audit-log config update --help",
			args:    "audit-log config update --help",
			fixture: "auditlog/confluent-audit-log-config-update-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestAuditLogConfigSpecSerialization() {
	original := LoadFixture(s.T(), "auditlogconfig-roundtrip-fixedpoint.golden")
	originalBytes := []byte(original)
	spec := mds.AuditLogConfigSpec{}
	if err := json.Unmarshal(originalBytes, &spec); err != nil {
		s.T().Fatal(err)
	}
	roundTripBytes, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		s.T().Fatal(err)
	}
	roundTrip := string(roundTripBytes)

	re := regexp.MustCompile(`[\r\n]+`)

	if re.ReplaceAllString(original, "") != re.ReplaceAllString(roundTrip, "") {
		s.T().Fail()
	}
}

func (s *CLITestSuite) TestAuditLogRoute() {
	tests := []CLITest{
		{
			name:    "confluent audit-log route list --help",
			args:    "audit-log route list --help",
			fixture: "auditlog/confluent-audit-log-route-list-help.golden",
		},
		{
			name:    "confluent audit-log route lookup --help",
			args:    "audit-log route lookup --help",
			fixture: "auditlog/confluent-audit-log-route-lookup-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runConfluentTest(tt)
	}
}

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
