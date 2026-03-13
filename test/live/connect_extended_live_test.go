//go:build live_test && (all || connect)

package live

import (
	"os"
	"testing"
)

func (s *CLILiveTestSuite) TestConnectCustomPluginCRUDLive() {
	t := s.T()
	t.Parallel()

	envID := os.Getenv("LIVE_TEST_ENVIRONMENT_ID")
	if envID == "" {
		t.Skip("Skipping: LIVE_TEST_ENVIRONMENT_ID must be set")
	}

	state := s.setupTestContext(t)

	// Register cleanup
	s.registerCleanup(t, "connect custom-plugin delete {{.plugin_id}} --force", state)

	steps := []CLILiveTest{
		{
			Name: "Use environment",
			Args: "environment use " + envID,
		},
		{
			Name: "List custom plugins",
			Args: "connect custom-plugin list",
		},
		{
			Name: "List custom runtimes",
			Args: "connect custom-runtime list --environment " + envID,
		},
	}

	for _, step := range steps {
		t.Run(step.Name, func(t *testing.T) {
			s.runLiveCommand(t, step, state)
		})
	}
}
