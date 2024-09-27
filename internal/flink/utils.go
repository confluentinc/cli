package flink

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

const cmfEnvironmentConfigPath = "~/.cmf/environment"

// getEnvironment returns the environment from the flag or the config file.
func getEnvironment(cmd *cobra.Command) string {
	// Check if someone has passed the environment flag
	environment, err := cmd.Flags().GetString("environment")
	if err == nil && environment != "" {
		return environment
	}
	// check if the environment is set in the config file
	cmfConfigFilePath := expandHomeDir(cmfEnvironmentConfigPath)
	if _, err := os.Stat(cmfConfigFilePath); os.IsNotExist(err) {
		return ""
	}
	data, err := ioutil.ReadFile(cmfConfigFilePath)
	if err != nil {
		return ""
	}
	fmt.Printf("Using environment from config file: %s\n", string(data))
	return strings.TrimSpace(string(data))
}

func expandHomeDir(path string) string {
	if strings.HasPrefix(path, "~") {
		if home, err := os.UserHomeDir(); err == nil {
			return strings.Replace(path, "~", home, 1)
		}
	}

	return path
}
