package kafka

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"

	// Import the official SDK types
	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
)

func (c *shareCommand) newGroupListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share groups.",
		Args:  cobra.NoArgs,
		RunE:  c.groupList,
	}

	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *shareCommand) groupList(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	shareGroups, err := kafkaREST.CloudClient.ListKafkaShareGroups()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, shareGroup := range shareGroups {
		list.Add(&shareGroupListOut{
			Cluster:       shareGroup.GetClusterId(),
			ShareGroup:    shareGroup.GetShareGroupId(),
			Coordinator:   getStringBrokerFromShareGroup(shareGroup),
			State:         shareGroup.GetState(),
			ConsumerCount: shareGroup.GetConsumerCount(),
		})
	}
	return list.Print()
}

// Helper function to extract broker information from share group
func getStringBrokerFromShareGroup(shareGroup interface{}) string {
	// Cast to the actual SDK type - it's v3.ShareGroupData, not a pointer
	if sg, ok := shareGroup.(kafkarestv3.ShareGroupData); ok {
		coordinator := sg.GetCoordinator()

		// GetCoordinator() returns a Relationship struct, not a pointer
		relationship := coordinator.GetRelated()

		// relationship will look like ".../v3/clusters/{cluster_id}/brokers/{broker_id}"
		splitString := strings.SplitAfter(relationship, "brokers/")
		// if relationship was an empty string or did not contain "brokers/"
		if len(splitString) < 2 {
			return ""
		}
		// returning brokerId
		return splitString[1]
	}

	// If type assertion fails, return N/A for now
	return "N/A"
}
