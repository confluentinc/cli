package flink

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentDescribe,
	}

	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.`)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentDescribe(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the environment to be retrieved
	environmentName := args[0]
	environment, err := cmfClient.DescribeEnvironment(cmd.Context(), environmentName)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	var defaultsBytes []byte
	defaultsBytes, err = json.Marshal(environment.FlinkApplicationDefaults)
	if err != nil {
		return fmt.Errorf(`failed to marshal defaults for environment "%s": %s`, environmentName, err)
	}

	table.Add(&flinkEnvironmentOutput{
		Name:                     environment.Name,
		KubernetesNamespace:      environment.KubernetesNamespace,
		FlinkApplicationDefaults: string(defaultsBytes),
		CreatedTime:              environment.CreatedTime.String(),
		UpdatedTime:              environment.UpdatedTime.String(),
	})
	return table.Print()
}
