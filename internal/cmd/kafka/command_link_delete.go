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
		//RunE:  c.delete,
	}

	if c.cfg.IsCloudLogin() {
		cmd.RunE = c.delete
		pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.RunE = c.deleteOnPrem
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	kafkaREST, err := c.GetCloudKafkaREST()
	if err != nil {
		return err
	}
	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	clusterId := kafkaClusterConfig.ID

	if httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.DeleteKafkaLink(kafkaREST.Context, clusterId, linkName).Execute(); err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	return nil
}
