//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestAuditLogLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name: "Describe audit log config",
			Args: "audit-log describe -o json",
			JSONFieldsExist: []string{"cluster_id"},
		},
		{
			Name: "List audit log routes",
			Args: "audit-log route list",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
