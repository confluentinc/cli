//go:build live_test && (all || core)

package live

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestApiKeyCRUDLive() {
	t := s.T()
	t.Parallel()
	state := s.setupTestContext(t)

	saName := uniqueName("apikey-sa")
	apiKeyDescription := "Live test API key"
	updatedDescription := "Updated live test API key"

	// Cleanup in LIFO order: delete API key first, then service account
	s.registerCleanup(t, "iam service-account delete {{.sa_id}} --force", state)
	s.registerCleanup(t, "api-key delete {{.api_key_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name:     "Create service account for API key",
			Args:     `iam service-account create ` + saName + ` --description "SA for API key live test" -o json`,
			ExitCode: 0,
			JSONFieldsExist: []string{"id"},
			CaptureID:       "sa_id",
		},
		{
			Name:            "Create API key",
			Args:            `api-key create --resource cloud --service-account {{.sa_id}} --description "` + apiKeyDescription + `" -o json`,
			UseStateVars:    true,
			ExitCode:        0,
			JSONFieldsExist: []string{"api_key", "api_secret"},
			WantFunc: func(t *testing.T, output string, state *LiveTestState) {
				t.Helper()
				id := extractJSONField(t, output, "api_key")
				require.NotEmpty(t, id, "failed to extract api_key from output:\n%s", output)
				state.Set("api_key_id", id)
				t.Logf("Captured api_key_id = %s", id)
			},
		},
		{
			Name:         "List API keys",
			Args:         "api-key list --service-account {{.sa_id}}",
			UseStateVars: true,
			ExitCode:     0,
			WantFunc: func(t *testing.T, output string, state *LiveTestState) {
				t.Helper()
				apiKeyID := state.Get("api_key_id")
				if apiKeyID != "" {
					require.Contains(t, output, apiKeyID, "API key list should contain the created key")
				}
			},
		},
		{
			Name:         "Update API key description",
			Args:         `api-key update {{.api_key_id}} --description "` + updatedDescription + `"`,
			UseStateVars: true,
			ExitCode:     0,
		},
		{
			Name:         "Describe updated API key",
			Args:         "api-key describe {{.api_key_id}} -o json",
			UseStateVars: true,
			ExitCode:     0,
			JSONFields: map[string]string{
				"description": updatedDescription,
			},
			JSONFieldsExist: []string{"key", "owner"},
		},
		{
			Name:         "Delete API key",
			Args:         "api-key delete {{.api_key_id}} --force",
			UseStateVars: true,
			ExitCode:     0,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
