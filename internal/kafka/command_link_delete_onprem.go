package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *linkCommand) newDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link>",
		Short: "Delete a cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.deleteOnPrem,
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *linkCommand) deleteOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	if httpResp, err := client.ClusterLinkingV3Api.DeleteKafkaLink(ctx, clusterId, linkName, nil); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	output.Printf(errors.DeletedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
