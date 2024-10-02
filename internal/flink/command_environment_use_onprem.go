package flink

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

func (c *command) newEnvironmentUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Use an environment as default for Flink applications",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentUse,
	}

	return cmd
}

func (c *command) environmentUse(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment := args[0]

	// Check if the environment exists or not
	_, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), environment)
	if httpResponse != nil && httpResponse.StatusCode != 200 {
		return fmt.Errorf("environment \"%s\" does not exist", environment)
	}

	cmfConfigFilePath := expandHomeDir(cmfEnvironmentConfigPath)
	configWritten := true
	// See if the file exists or not
	if _, err := os.Stat(cmfConfigFilePath); os.IsNotExist(err) {
		// try to create the file
		if err := os.MkdirAll(filepath.Dir(cmfConfigFilePath), 0755); err != nil {
			// if you failed to create the file, save the environment in an environment variable
			configWritten = false
		} else {
			// create the file and write the environment in it
			if err := os.WriteFile(cmfConfigFilePath, []byte(environment), 0644); err != nil {
				configWritten = false
			}
		}
	} else {
		// if the file exists, write the environment in it
		if err := os.WriteFile(cmfConfigFilePath, []byte(environment), 0644); err != nil {
			configWritten = false
		}
	}
	if !configWritten {
		return fmt.Errorf("failed to set the environment \"%s\" as default", environment)
	}
	output.Printf(false, "Environment \"%s\" is set as default", environment)
	return nil
}
