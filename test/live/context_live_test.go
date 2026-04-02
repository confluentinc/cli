//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestContextAndConfigurationLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	steps := []CLILiveTest{
		{
			Name:     "List contexts",
			Args:     "context list",
			Contains: []string{"login-"},
		},
		{
			Name: "Describe current context",
			Args: "context describe -o json",
			JSONFieldsExist: []string{"name", "platform"},
		},
		{
			Name: "List configuration",
			Args: "configuration list",
		},
		{
			Name: "Describe disable_updates config",
			Args: "configuration describe disable_updates -o json",
			JSONFieldsExist: []string{"name", "value"},
		},
		{
			Name: "Update disable_update_check config",
			Args: "configuration update disable_update_check true",
		},
		{
			Name: "Verify config updated",
			Args: "configuration describe disable_update_check -o json",
			JSONFields: map[string]string{
				"value": "true",
			},
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
