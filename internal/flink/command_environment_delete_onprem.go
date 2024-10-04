package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newEnvironmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink environments.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.environmentDelete,
	}

	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CERT_AUTHORITY_PATH" may be set in place of this flag.`)

	pcmd.AddForceFlag(cmd)

	return cmd
}

func (c *command) environmentDelete(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := cmfClient.DescribeEnvironment(cmd.Context(), name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkEnvironment); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		return cmfClient.DeleteEnvironment(cmd.Context(), name)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkEnvironment)
	return err
}
