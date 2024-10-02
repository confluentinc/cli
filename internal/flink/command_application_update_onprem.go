package flink

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newApplicationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationUpdate,
	}

	cmd.Flags().String("environment", "", "Name of the Environment to update the FlinkApplication in.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *command) applicationUpdate(cmd *cobra.Command, args []string) error {
	environment, err := getEnvironment(cmd)
	if err != nil {
		return err
	}
	// Disallow human output for this command
	if output.GetFormat(cmd) == output.Human {
		return errors.NewErrorWithSuggestions("human output is not supported for this command", "Try using --output flag with json or yaml.\n")
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Check if the application already exists
	resourceFilePath := args[0]
	// Read file contents
	data, err := os.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var application cmfsdk.Application
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &application)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &application)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		return err
	}

	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, _ := cmfClient.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf(`application "%s" does not exist in the environment "%s"`, applicationName, environment)
	}

	outputApplication, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateApplication(cmd.Context(), environment, application)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to update application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}

	return output.SerializedOutput(cmd, outputApplication)
}
