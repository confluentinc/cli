//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestOrganizationLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name:     "List organizations",
			Args:     "organization list -o json",
			JSONFieldsExist: []string{"id"},
		},
		{
			Name:     "Describe organization",
			Args:     "organization describe -o json",
			JSONFieldsExist: []string{"id", "name"},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
