//go:build live_test && (all || iam)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestIAMUserLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name: "List users",
			Args: "iam user list -o json",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name: "List user invitations",
			Args: "iam user invitation list",
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
