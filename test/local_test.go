package test

import "time"

func (s *CLITestSuite) TestLocal() {
	tests := []CLITest{
		{args: "local kafka stop", fixture: "local/stop.golden"},
		{args: "local kafka start", fixture: "local/start.golden", regex: true},
	}

	for _, tt := range tests {
		tt.workflow = true
		tt.login = "platform"
		s.runIntegrationTest(tt)
	}

	time.Sleep(120 * time.Second)

	tests2 := []CLITest{
		{args: "local kafka topic create test", fixture: "local/topic_create.golden"},
		{args: "local kafka topic list", fixture: "local/topic_list.golden"},
		{args: "local kafka topic delete test --force", fixture: "local/topic_delete.golden"},
		{args: "local kafka stop", fixture: "local/stop.golden"},
	}

	for _, tt := range tests2 {
		tt.workflow = true
		tt.login = "platform"
		s.runIntegrationTest(tt)
	}
}
