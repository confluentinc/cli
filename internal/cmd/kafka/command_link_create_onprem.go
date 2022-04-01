package kafka

import (
	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (c *linkCommand) createOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]
	bootstrapServer, err := cmd.Flags().GetString(destinationBootstrapServerFlagName)
	if err != nil {
		return err
	}
	destinationClusterId, err := cmd.Flags().GetString(destinationClusterIdFlagName)
	if err != nil {
		return nil
	}
	configMap, err := c.parseConfigMap(bootstrapServer)

	restClient, restContext, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(restClient, restContext)
	if err != nil {
		return err
	}
	data := kafkarestv3.CreateLinkRequestData{Configs: toCPCreateTopicConfigs(configMap)}
	data.DestinationClusterId = destinationClusterId
	opts := &kafkarestv3.CreateKafkaLinkOpts{CreateLinkRequestData: optional.NewInterface(data)}

	if httpResp, err := restClient.ClusterLinkingV3Api.CreateKafkaLink(restContext, clusterId, linkName, opts); err != nil {
		return kafkaRestError(pcmd.GetCPKafkaRestBaseUrl(restClient), err, httpResp)
	}

	utils.Printf(cmd, errors.CreatedLinkMsg, linkName)
	return nil
}