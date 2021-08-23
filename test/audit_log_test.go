package test

import (
	"encoding/json"
	"fmt"
	"regexp"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

func (s *CLITestSuite) TestAuditLogConfig() {
	tests := []CLITest{
		{
			args:    "audit-log config describe --help",
			fixture: "auditlog/config-describe-help.golden",
		},
		{
			args:    "audit-log config edit --help",
			fixture: "auditlog/config-edit-help.golden",
		},
		{
			args:    "audit-log config update --help",
			fixture: "auditlog/config-update-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runOnPremTest(tt)
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
			args:    "audit-log route list --help",
			fixture: "auditlog/route-list-help.golden",
		},
		{
			args:    "audit-log route lookup --help",
			fixture: "auditlog/route-lookup-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "default"
		s.runOnPremTest(tt)
	}
}

func (s *CLITestSuite) TestAuditConfigMigrate() {
	migration1 := GetInputFixturePath(s.T(), "auditlog", "config-migration-server1.golden")
	migration2 := GetInputFixturePath(s.T(), "auditlog", "config-migration-server2.golden")

	malformed := GetInputFixturePath(s.T(), "auditlog", "malformed-migration.golden")
	nullFields := GetInputFixturePath(s.T(), "auditlog", "null-fields-migration.golden")

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
		s.runOnPremTest(tt)
	}
}
