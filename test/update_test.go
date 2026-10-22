package test

func (s *CLITestSuite) TestUpdate() {
	// Tests which accept the update replace the running binary, so each one needs its own
	// copy of it; declining the update leaves the binary alone.
	tests := []CLITest{
		{args: "update", fixture: "update/update.golden", input: "y\n", isolatedBin: true},
		{args: "update", fixture: "update/update-no.golden", input: "n\n"},
		{args: "update --major", fixture: "update/update-major.golden", input: "y\n", isolatedBin: true},
		{args: "update --no-verify", fixture: "update/update.golden", input: "y\n", isolatedBin: true},
		{args: "update --yes", fixture: "update/update-yes.golden", isolatedBin: true},
	}

	for _, test := range tests {
		s.runIntegrationTest(test)
	}
}
