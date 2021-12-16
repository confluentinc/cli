package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *linkCommand) newDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <link>",
		Short: "Delete a previously created cluster link.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
	}
}

func (c *linkCommand) delete(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameDelete(kafkaREST.Context, lkc, linkName)
	if err == nil {
		utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	}

	return handleOpenApiError(httpResp, err, kafkaREST)
}
