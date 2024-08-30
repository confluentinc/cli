package flink

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newConnectionCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink connection.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.connectionCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create Flink connection "my-connection" in AWS us-west-2 for openai with endpoint and api key`,
				Code: "confluent flink connection create my-connection --cloud aws --region us-west-2 " +
					"--connectionType openai --endpoint https://api.openai.com/v1/chat/completions --api_key mykey",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("type", "", "Specify the connection type. Supported types are: "+strings.Join(supportedConnectionTypes(), ", "))
	cmd.Flags().String("endpoint", "", "Specify endpoint for the connection")
	AddConnectionSecretFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("type"))
	cobra.CheckErr(cmd.MarkFlagRequired("endpoint"))
	AddConnectionSecretFlagChecks(cmd)

	return cmd
}

func (c *command) connectionCreate(cmd *cobra.Command, args []string) error {
	// TODO
	return nil
}
