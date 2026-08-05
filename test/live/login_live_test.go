//go:build live_test && (all || auth)

package live

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func (s *CLILiveTestSuite) TestLoginLogoutLive() {
	t := s.T()
	t.Parallel()

	email := requiredEnv(t, "CONFLUENT_CLOUD_EMAIL")
	password := requiredEnv(t, "CONFLUENT_CLOUD_PASSWORD")

	// Create isolated HOME directory (not using setupTestContext since that already logs in)
	homeDir, err := os.MkdirTemp("", "cli-live-auth-*")
	require.NoError(t, err)
	t.Cleanup(func() { os.RemoveAll(homeDir) })

	homeEnv := homeEnvVar(homeDir)

	// Clear all cloud auth env vars so the CLI can't auth implicitly — forces explicit login
	noAuthEnv := []string{
		homeEnv,
		"CONFLUENT_CLOUD_API_KEY=",
		"CONFLUENT_CLOUD_API_SECRET=",
		"CONFLUENT_CLOUD_EMAIL=",
		"CONFLUENT_CLOUD_PASSWORD=",
	}

	// Step 1: Before login, commands requiring auth should fail
	t.Run("Verify not logged in", func(t *testing.T) {
		output := s.runRawCommand(t, "environment list", noAuthEnv, "", 1)
		t.Logf("Pre-login output: %s", output)
	})

	// Step 2: Login with email and password
	t.Run("Login with credentials", func(t *testing.T) {
		output := s.runRawCommand(t, "login", []string{
			homeEnv,
			"CONFLUENT_CLOUD_API_KEY=",
			"CONFLUENT_CLOUD_API_SECRET=",
			fmt.Sprintf("CONFLUENT_CLOUD_EMAIL=%s", email),
			fmt.Sprintf("CONFLUENT_CLOUD_PASSWORD=%s", password),
		}, "", 0)
		require.Contains(t, output, "Logged in", "login should succeed")
	})

	// Step 3: Verify login succeeded by running an authenticated command (without API key env vars)
	t.Run("Verify logged in", func(t *testing.T) {
		s.runRawCommand(t, "environment list", noAuthEnv, "", 0)
	})

	// Step 4: Logout
	t.Run("Logout", func(t *testing.T) {
		s.runRawCommand(t, "logout", noAuthEnv, "", 0)
	})

	// Step 5: After logout, commands requiring auth should fail again
	t.Run("Verify logged out", func(t *testing.T) {
		output := s.runRawCommand(t, "environment list", noAuthEnv, "", 1)
		t.Logf("Post-logout output: %s", output)
	})
}
