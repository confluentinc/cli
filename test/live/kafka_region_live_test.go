//go:build live_test && (all || kafka)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestKafkaRegionListLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name:     "List kafka regions for aws",
			Args:     "kafka region list --cloud aws",
			Contains: []string{"us-east-1"},
		},
		{
			Name:     "List kafka regions for gcp",
			Args:     "kafka region list --cloud gcp",
			Contains: []string{"us-east1"},
		},
		{
			Name:     "List kafka regions for azure",
			Args:     "kafka region list --cloud azure",
			Contains: []string{"eastus"},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
