package flink

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newApplicationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CERT_AUTHORITY_PATH" may be set in place of this flag.`)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	return cmd
}

func (c *command) applicationDescribe(cmd *cobra.Command, args []string) error {
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

	// Get the name of the application to be retrieved
	applicationName := args[0]
	cmfApplication, httpResponse, err := cmfClient.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to describe application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}

	table := output.NewTable(cmd)
	var metadataBytes, specBytes, statusBytes []byte
	metadataBytes, err = json.Marshal(cmfApplication.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %s", err)
	}
	specBytes, err = json.Marshal(cmfApplication.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %s", err)
	}
	statusBytes, err = json.Marshal(cmfApplication.Status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %s", err)
	}

	table.Add(&flinkApplicationOutput{
		ApiVersion: cmfApplication.ApiVersion,
		Kind:       cmfApplication.Kind,
		Metadata:   string(metadataBytes),
		Spec:       string(specBytes),
		Status:     string(statusBytes),
	})
	return table.Print()
}
