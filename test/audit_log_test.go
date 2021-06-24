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
			Name:    "confluent audit-log config describe --help",
			Args:    "audit-log config describe --help",
			Fixture: "auditlog/confluent-audit-log-config-describe-help.golden",
		},
		{
			Name:    "confluent audit-log config edit --help",
			Args:    "audit-log config edit --help",
			Fixture: "auditlog/confluent-audit-log-config-edit-help.golden",
		},
		{
			Name:    "confluent audit-log config update --help",
			Args:    "audit-log config update --help",
			Fixture: "auditlog/confluent-audit-log-config-update-help.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
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
			Name:    "confluent audit-log route list --help",
			Args:    "audit-log route list --help",
			Fixture: "auditlog/confluent-audit-log-route-list-help.golden",
		},
		{
			Name:    "confluent audit-log route lookup --help",
			Args:    "audit-log route lookup --help",
			Fixture: "auditlog/confluent-audit-log-route-lookup-help.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		s.RunConfluentTest(tt)
	}
}

func (s *CLITestSuite) TestAuditConfigMigrate() {
	migration1 := GetInputFixturePath(s.T(), "auditlog", "config-migration-server1.golden")
	migration2 := GetInputFixturePath(s.T(), "auditlog", "config-migration-server2.golden")

	malformed := GetInputFixturePath(s.T(), "auditlog", "malformed-migration.golden")
	nullFields := GetInputFixturePath(s.T(), "auditlog", "null-fields-migration.golden")

	tests := []CLITest{
		{
			Args: fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2 --bootstrap-servers new_bootstrap_1 --authority NEW.CRN.AUTHORITY.COM", migration1, migration2),
			Fixture: "auditlog/migration-result-with-warnings.golden",
		},
		{
			Args: fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s "+
				"--bootstrap-servers new_bootstrap_2", malformed, migration2),
			Contains: "Ignoring property file",
		},
		{
			Args:    fmt.Sprintf("audit-log migrate config --combine cluster123=%s,clusterABC=%s", nullFields, nullFields),
			Fixture: "auditlog/empty-migration-result.golden",
		},
	}

	for _, tt := range tests {
		tt.Login = "default"
		tt.Workflow = true
		s.RunConfluentTest(tt)
	}
}
