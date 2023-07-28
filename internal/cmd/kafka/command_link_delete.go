package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *linkCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <link>",
		Short:             "Delete a cluster link.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.ClusterLink, linkName, linkName)
	if _, err := form.ConfirmDeletion(cmd, promptMsg, linkName); err != nil {
		return err
	}

	if httpResp, err := kafkaREST.CloudClient.DeleteKafkaLink(cluster.ID, linkName); err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	output.Printf(errors.DeletedResourceMsg, resource.ClusterLink, linkName)
	return nil
}
