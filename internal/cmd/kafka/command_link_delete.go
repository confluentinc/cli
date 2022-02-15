package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *linkCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <link>",
		Short: "Delete a previously created cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}

	if c.cfg.IsOnPremLogin() {
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, clusterId, err := c.getKafkaRestComponents(cmd)
	if err != nil {
		return err
	}

	if httpResp, err := client.ClusterLinkingV3Api.DeleteKafkaLink(ctx, clusterId, linkName, nil); err != nil {
		return handleOpenApiError(httpResp, err, client)
	}

	utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	return nil
}
