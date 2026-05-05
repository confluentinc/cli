//go:build live_test && (all || rtce)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestRegionCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	// Variables
	regionName := uniqueName("region")

	// Cleanup (LIFO)

	steps := []CLILiveTest{
		{
			Name:         "List rtce regions",
			Args:         "rtce region list",
			UseStateVars: true,
			ExitCode:     0,
			Contains:     []string{regionName},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
