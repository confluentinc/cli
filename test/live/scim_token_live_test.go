//go:build live_test && (all || core)

package live

import (
	"testing"
)

func (s *CLILiveTestSuite) TestScimTokenCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	// Variables
	scimTokenName := uniqueName("scimto")

	// Cleanup (LIFO)
	s.registerCleanup(t, "org scim-token delete {{.scim_token_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name:            "Create org scim token",
			Args:            "org scim-token create " + scimTokenName + " -o json",
			ExitCode:        0,
			JSONFields:      map[string]string{},
			JSONFieldsExist: []string{"id"},
			CaptureID:       "scim_token_id",
		},
		{
			Name:         "List org scim tokens",
			Args:         "org scim-token list",
			UseStateVars: true,
			ExitCode:     0,
			Contains:     []string{scimTokenName},
		},
		{
			Name:         "Delete org scim token",
			Args:         "org scim-token delete {{.scim_token_id}} --force",
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Verify deletion",
			Args:         "org scim-token describe {{.scim_token_id}}",
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
