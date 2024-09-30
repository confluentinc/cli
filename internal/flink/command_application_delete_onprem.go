package flink

import (
	"errors"
	"io"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	perrors "github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

type deleteApplicationFailure struct {
	Application string `human:"Application" serialized:"application"`
	Environment string `human:"Environment" serialized:"environment"`
	Reason      string `human:"Reason" serialized:"reason"`
	StausCode   int    `human:"Status Code" serialized:"status_code"`
}

func (c *unauthenticatedCommand) newApplicationDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink Applications.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.applicationDelete,
	}

	cmd.Flags().String("environment", "", "Name of the Environment to get the FlinkApplication from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *unauthenticatedCommand) applicationDelete(cmd *cobra.Command, args []string) error {
	environment := getEnvironment(cmd)
	if environment == "" {
		return perrors.NewErrorWithSuggestions("environment name is required", "You can use the --environment flag or set the default environment using `confluent flink environment use <name>` command")
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		httpResp, err := cmfClient.DefaultApi.DeleteApplication(cmd.Context(), environment, name)
		if err != nil && httpResp != nil {
			if httpResp.Body != nil {
				defer httpResp.Body.Close()
				respBody, parseError := io.ReadAll(httpResp.Body)
				if parseError == nil {
					return errors.New(string(respBody))
				}
			}
		}
		return err
	}

	_, err = deletion.Delete(args, deleteFunc, resource.OnPremFlinkApplication)
	return err
}
