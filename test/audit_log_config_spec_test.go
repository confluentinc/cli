package test

import (
	"encoding/json"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

func (s *CLITestSuite) TestAuditLogConfigSpecSerialization() {
	original := loadFixture(s.T(), "auditlogconfig-roundtrip-fixedpoint.golden")
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
	if original != roundTrip {
		s.T().Fail()
	}
}
