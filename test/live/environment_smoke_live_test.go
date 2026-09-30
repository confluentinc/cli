//go:build live_test && (all || smoke)

package live

import (
	"testing"
)

// TestEnvironmentSmokeLive is the lightweight liveness probe run on the smoke schedule
// (every 3 hours, per platform). It verifies CLI binary works, login succeeds, and the
// org/v2/environments API returns a parseable JSON array. A failure here means the CLI
// for this platform cannot complete the most basic authenticated read against
// Confluent Cloud — pages live in the smoke metric, not in this test's output.
func (s *CLILiveTestSuite) TestEnvironmentSmokeLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name:            "List environments",
			Args:            "environment list -o json",
			ExitCode:        0,
			JSONFieldsExist: []string{"id", "name"},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
