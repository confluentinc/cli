package flink

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func (c *unauthenticatedCommand) newApplicationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationUpdate,
	}

	cmd.Flags().StringP("environment", "e", "", "Name of the Environment to get the FlinkApplication from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")

	cmd.MarkFlagRequired("environment")

	return cmd
}

func (c *unauthenticatedCommand) applicationUpdate(cmd *cobra.Command, args []string) error {
	environmentName, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentName == "" {
		fmt.Errorf("Environment name is required")
		return nil
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Check if the application already exists
	resourceFilePath := args[0]
	// Read file contents
	data, err := ioutil.ReadFile(resourceFilePath)
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
	_, httpResponse, err := cmfClient.DefaultApi.GetApplication(cmd.Context(), environmentName, applicationName, nil)
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode != 200 {
		return fmt.Errorf("application \"%s\" does not exist in the environment \"%s\"", applicationName, environmentName)
	}

	outputApplication, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateApplication(cmd.Context(), environmentName, application)
	defer httpResponse.Body.Close()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			respBody, parseError := ioutil.ReadAll(httpResponse.Body)
			if httpResponse.Body != nil {
				if parseError == nil {
					return fmt.Errorf("failed to update application \"%s\" in the environment \"%s\": %s", applicationName, environmentName, string(respBody))
				}
			}
		}
		return fmt.Errorf("failed to update application \"%s\" in the environment \"%s\": %s", applicationName, environmentName, err)
	}
	// TODO: can err == nil and status code non-20x?

	if output.GetFormat(cmd) == output.Human {
		// TODO: Add different output formats
	}
	return output.SerializedOutput(cmd, outputApplication)
}
