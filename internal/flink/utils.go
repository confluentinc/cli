package flink

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	perrors "github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

const cmfEnvironmentConfigPath = "~/.cmf/environment"

// getEnvironment returns the environment from the flag or the config file.
func getEnvironment(cmd *cobra.Command) (string, error) {
	envEmptyErr := perrors.NewErrorWithSuggestions("environment is required.", "Specify an environment with `--environment` or set the default environment using `confluent flink environment use`.")

	// Check if someone has passed the environment flag
	environment, err := cmd.Flags().GetString("environment")
	if err == nil && environment != "" {
		return environment, nil
	}
	// check if the environment is set in the config file
	cmfConfigFilePath := expandHomeDir(cmfEnvironmentConfigPath)
	if _, err := os.Stat(cmfConfigFilePath); os.IsNotExist(err) {
		// Don't return the "ErrNotExist" error as it's not relevant, just return the empty environment error.
		return "", envEmptyErr
	}
	data, err := os.ReadFile(cmfConfigFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to read environment from config file %s: %w", cmfConfigFilePath, err)
	}

	trimmedEnv := strings.TrimSpace(string(data))
	if trimmedEnv == "" {
		return "", envEmptyErr
	}
	output.Printf(false, "Using environment from config file: %s\n", trimmedEnv)
	return trimmedEnv, nil
}

func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}

	return path
}
