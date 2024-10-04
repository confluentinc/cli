package flink

import (
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

	applicationName := args[0]
	application, err := cmfClient.DescribeApplication(cmd.Context(), environment, applicationName)
	if err != nil {
		return err
	}

	return output.SerializedOutput(cmd, application)
}
