//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestEnvironmentCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	envName := uniqueName("env")
	updatedEnvName := envName + "-updated"

	// Register cleanup immediately — runs even if test fails (LIFO order)
	s.registerCleanup(t, "environment delete {{.env_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name:     "Create environment",
			Args:     "environment create " + envName + " -o json",
			ExitCode: 0,
			JSONFields: map[string]string{
				"name": envName,
			},
			JSONFieldsExist: []string{"id"},
			CaptureID:       "env_id",
		},
		{
			Name:         "Describe environment",
			Args:         "environment describe {{.env_id}} -o json",
			UseStateVars: true,
			ExitCode:     0,
			JSONFields: map[string]string{
				"name": envName,
				"id":   "", // any non-empty value
			},
		},
		{
			Name:         "List environments",
			Args:         "environment list",
			UseStateVars: true,
			ExitCode:     0,
			Contains:     []string{envName},
		},
		{
			Name:         "Update environment name",
			Args:         "environment update {{.env_id}} --name " + updatedEnvName,
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Describe updated environment",
			Args:         "environment describe {{.env_id}} -o json",
			UseStateVars: true,
			ExitCode:     0,
			JSONFields: map[string]string{
				"name": updatedEnvName,
			},
		},
		{
			Name:         "Delete environment",
			Args:         "environment delete {{.env_id}} --force",
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Verify deletion",
			Args:         "environment describe {{.env_id}}",
			UseStateVars: true,
			ExitCode:     1,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
