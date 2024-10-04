package flink

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newEnvironmentUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <name>",
		Short: "Use an environment as default for Flink applications",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentUse,
	}

	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CERT_AUTHORITY_PATH" may be set in place of this flag.`)

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
	if err != nil || (httpResponse != nil && httpResponse.StatusCode != http.StatusOK) {
		return fmt.Errorf(`environment "%s" does not exist`, environment)
	}

	cmfConfigFilePath := expandHomeDir(cmfEnvironmentConfigPath)
	var fileWriteErr error

	// See if the file exists or not
	if _, err := os.Stat(cmfConfigFilePath); os.IsNotExist(err) {
		// Try to create the directory.
		if fileWriteErr = os.MkdirAll(filepath.Dir(cmfConfigFilePath), 0755); fileWriteErr == nil {
			// Create the file and write the environment in it
			fileWriteErr = os.WriteFile(cmfConfigFilePath, []byte(environment), 0644)
		}
	} else {
		// If the file exists, write the environment in it.
		fileWriteErr = os.WriteFile(cmfConfigFilePath, []byte(environment), 0644)
	}
	if fileWriteErr != nil {
		return fmt.Errorf(`failed to set the environment "%s" as default, couldn't write to file "%s": %w`, environment, cmfConfigFilePath, fileWriteErr)
	}

	output.Printf(false, `Environment "%s" is set as default`, environment)
	return nil
}
