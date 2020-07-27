package test

// Example integration tests for example command config file show <num-times>
// This uses the test binary and hits the test server and verifies outputs (difference from system test?)

// run with: make test INT_TEST_ARGS="-run TestCLI/TestFileShow"
// Check cmd/lint/main.go if lint fails on lint-cli
func (s *CLITestSuite) TestFileShow() {
	defer s.destroy() // TODO: this doesn't work

	tests := []CLITest{ // create a series of internal tests that require command to match expected output
		// expected output are placed at test/fixtures/output
		{name: "succeed if showing existing config filepath once", args: "config file show 1", fixture: "config-file-show-1.golden"},
		{name: "succeed if showing existing config filepath multiple times via argument", args: "config file show 5", fixture: "config-file-show-5.golden"},
	}
	resetConfiguration(s.T(), "ccloud")
	for _, tt := range tests { // for each test in tests
		if tt.name == "" { // set name to args if empty
			tt.name = tt.args
		}
		tt.workflow = true // whether want to reset config/state between commandline actions
		// kafkaAPIURL := serveKafkaAPI(s.T()).URL // find URL of kafka cloud?
		// s.runCcloudTest(tt, serve(s.T(), kafkaAPIURL).URL, kafkaAPIURL) // run tests against cloud
	}
}
