//go:build live_test && (all || billing)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestBillingLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name: "List billing prices",
			Args: "billing price list",
		},
		{
			Name: "List billing costs",
			Args: "billing cost list",
		},
		{
			Name: "List billing promos",
			Args: "billing promo list",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
