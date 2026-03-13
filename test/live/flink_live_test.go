//go:build live_test && (all || flink)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestFlinkRegionListLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name:     "List flink regions for aws",
			Args:     "flink region list --cloud aws",
			Contains: []string{"us-east-1"},
		},
		{
			Name:     "List flink regions for gcp",
			Args:     "flink region list --cloud gcp",
			Contains: []string{"us-east1"},
		},
		{
			Name:     "List flink regions for azure",
			Args:     "flink region list --cloud azure",
			Contains: []string{"eastus"},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
