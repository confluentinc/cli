package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newEnvironmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink environments.",
		Args:  cobra.NoArgs,
		RunE:  c.environmentList,
	}

	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CERT_AUTHORITY_PATH" may be set in place of this flag.`)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentList(cmd *cobra.Command, _ []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environments, err := cmfClient.ListEnvironments(cmd.Context())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, env := range environments {
			list.Add(&flinkEnvironmentSummary{
				Name:        env.Name,
				CreatedTime: env.CreatedTime.String(),
				UpdatedTime: env.UpdatedTime.String(),
			})
		}
		return list.Print()
	}
	return output.SerializedOutput(cmd, environments)
}
