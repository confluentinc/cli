package flink

import (
	"strings"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *command) newConnectionUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update [name]",
		Short:             "Update a Flink connection. Only secret can be updated",
		Args:              cobra.MaximumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs), // TODO
		RunE:              c.connectionUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update api key of a Flink connection.`,
				Code: `confluent flink connection update my-connection --cloud aws --region us-west-2 --api-key new-key`,
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("type", "", "Specify the connection type. Supported types are: "+strings.Join(supportedConnectionTypes(), ", "))
	AddConnectionSecretFlags(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))
	cobra.CheckErr(cmd.MarkFlagRequired("type"))
	AddConnectionSecretFlagChecks(cmd)

	return cmd
}

func (c *command) connectionUpdate(cmd *cobra.Command, args []string) error {
	// TODO
	return nil
}
