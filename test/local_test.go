package test

import (
	"runtime"
	"time"
)

func (s *CLITestSuite) TestLocalKafka() {
	if runtime.GOOS == "darwin" {
		s.T().Skip()
	}

	tests := []CLITest{
		{args: "local kafka stop", fixture: "local/kafka/stop-empty.golden"},
		{args: "local kafka start", fixture: "local/kafka/start.golden", regex: true},
	}

	for _, tt := range tests {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}

	time.Sleep(25 * time.Second)

	tests2 := []CLITest{
		{args: "local kafka topic create test", fixture: "local/kafka/topic/create.golden"},
		{args: "local kafka topic list", fixture: "local/kafka/topic/list.golden"},
		{args: "local kafka topic describe test", fixture: "local/kafka/topic/describe.golden"},
		{args: "local kafka topic delete test --force", fixture: "local/kafka/topic/delete.golden"},
		{args: "local kafka stop", fixture: "local/kafka/stop.golden", regex: true},
	}

	for _, tt := range tests2 {
		tt.workflow = true
		s.runIntegrationTest(tt)
	}
}
