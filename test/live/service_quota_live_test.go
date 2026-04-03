//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestServiceQuotaLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name: "List service quotas",
			Args: "service-quota list organization",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
