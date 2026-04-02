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
			Name:            "Describe enable_color config",
			Args:            "configuration describe enable_color -o json",
			JSONFieldsExist: []string{"name", "value"},
		},
		{
			Name: "Update enable_color config",
			Args: "configuration update enable_color true",
		},
		{
			Name: "Verify config updated",
			Args: "configuration describe enable_color -o json",
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
