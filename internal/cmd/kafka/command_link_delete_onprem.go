package kafka

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
)

func (c *linkCommand) deleteOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
	if err != nil {
		return err
	}

	if httpResp, err := client.ClusterLinkingV3Api.DeleteKafkaLink(ctx, clusterId, linkName, nil); err != nil {
		return kafkaRestError(pcmd.GetCPKafkaRestBaseUrl(client), err, httpResp)
	}

	utils.Printf(cmd, errors.DeletedLinkMsg, linkName)
	return nil
}
