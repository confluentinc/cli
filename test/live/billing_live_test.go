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
			Args: "billing price list --cloud aws --region us-east-1",
		},
		{
			Name: "List billing costs",
			Args: "billing cost list --start-date 2026-03-01 --end-date 2026-03-31",
		},
		{
			Name: "List billing promos",
			Args: "billing promo list",
		},
		{
			Name: "Describe billing payment",
			Args: "billing payment describe",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
