package test

import (
	"encoding/json"
	"fmt"
	"regexp"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
)

func (s *CLITestSuite) TestAuditLogDescribe() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe.golden"})
}

func (s *CLITestSuite) TestAuditLogConfig() {
	tests := []CLITest{
		{
			name:    "confluent audit-log config describe --help",
			args:    "audit-log config describe --help",
			fixture: "audit-log/config/describe-help.golden",
		},
		{
			name:    "confluent audit-log config edit --help",
			args:    "audit-log config edit --help",
			fixture: "audit-log/config/edit-help.golden",
		},
		{
			name:    "confluent audit-log config update --help",
			args:    "audit-log config update --help",
			fixture: "audit-log/config/update-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestAuditLogConfigSpecSerialization() {
	original := LoadFixture(s.T(), "audit-log/config/roundtrip-fixedpoint.golden")
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
			fixture: "audit-log/route/list-help.golden",
		},
		{
			name:    "confluent audit-log route lookup --help",
			args:    "audit-log route lookup --help",
			fixture: "audit-log/route/lookup-help.golden",
		},
	}

	for _, tt := range tests {
		tt.login = "cloud"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestAuditConfigMigrate() {
	migration1 := GetInputFixturePath(s.T(), "audit-log", "config-migration-server1.golden")
	migration2 := GetInputFixturePath(s.T(), "audit-log", "config-migration-server2.golden")

	malformed := GetInputFixturePath(s.T(), "audit-log", "malformed-migration.golden")
	nullFields := GetInputFixturePath(s.T(), "audit-log", "null-fields-migration.golden")

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
		tt.login = "platform"
		s.runIntegrationTest(tt)
	}
}

func (s *CLITestSuite) TestAuditLogDisabledDescribe() {
	s.runIntegrationTest(CLITest{args: "audit-log describe", login: "cloud", fixture: "audit-log/describe-fail.golden", disableAuditLog: true, wantErrCode: 1})
}
