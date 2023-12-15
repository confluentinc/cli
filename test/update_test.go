package test

import (
	"os"
	"path/filepath"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func (s *CLITestSuite) TestUpdate() {
	s.T().Skip("Skipping this test until its less flaky")

	// Remove the cache file so we'll see the update prompt
	home, err := os.UserHomeDir()
	require.NoError(s.T(), err)
	path := filepath.Join(home, ".confluent", "update_check")
	err = os.RemoveAll(path) // RemoveAll so we don't return an error if file doesn't exist
	require.NoError(s.T(), err)

	// Be nice and restore the config when we're done
	configFile := config.GetDefaultFilename()

	oldConfig, err := os.ReadFile(configFile)
	require.NoError(s.T(), err)
	defer func() {
		err := os.WriteFile(configFile, oldConfig, 0600)
		require.NoError(s.T(), err)
	}()

	// Reset the config to a known empty state
	err = os.WriteFile(configFile, []byte(`{}`), 0600)
	require.NoError(s.T(), err)

	tests := []CLITest{
		{args: "version", fixture: "update/1.golden", regex: true},
		{args: "--help", contains: "Update the confluent CLI."},
		{name: "HACK: disable update checks"},
		{args: "version", fixture: "update/2.golden", regex: true},
		{args: "--help", contains: "Update the confluent CLI."},
		{name: "HACK: enabled checks, disable updates"},
		{args: "version", fixture: "update/2.golden", regex: true},
		{args: "--help", notContains: "Update the confluent CLI."},
	}

	for _, test := range tests {
		test.workflow = true
		switch test.name {
		case "HACK: disable update checks":
			err = os.WriteFile(configFile, []byte(`{"disable_update_checks": true}`), os.ModePerm)
			require.NoError(s.T(), err)
		case "HACK: enabled checks, disable updates":
			err = os.WriteFile(configFile, []byte(`{"disable_updates": true}`), os.ModePerm)
			require.NoError(s.T(), err)
		default:
			s.runIntegrationTest(test)
			if test.fixture == "update/1.golden" {
				// Remove the cache file so it _would_ prompt again (if not disabled)
				err = os.RemoveAll(path) // RemoveAll so we don't return an error if file doesn't exist
				require.NoError(s.T(), err)
			}
		}
	}
}
