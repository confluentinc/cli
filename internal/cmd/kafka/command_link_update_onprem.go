package kafka

import (
	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"
)

func (c *linkCommand) updateOnPrem(cmd *cobra.Command, args []string) error {
	linkName := args[0]

	configFile, err := cmd.Flags().GetString(configFileFlagName)
	if err != nil {
		return err
	}

	configMap := make(map[string]string)
	if configFile != "" {
		configMap, err = properties.FileToMap(configFile)
		if err != nil {
			return err
		}
	}

	if len(configMap) == 0 {
		return errors.New(errors.EmptyConfigErrorMsg)
	}

	client, ctx, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	clusterId, err := getClusterIdForRestRequests(client, ctx)
	if err != nil {
		return err
	}

	opts := &kafkarestv3.UpdateKafkaLinkConfigBatchOpts{
		AlterConfigBatchRequestData: optional.NewInterface(kafkarestv3.AlterConfigBatchRequestData{
			Data: toAlterConfigBatchRequestData(configMap),
		}),
	}

	if httpResp, err := client.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(ctx, clusterId, linkName, opts); err != nil {
		return kafkaRestError(pcmd.GetCPKafkaRestBaseUrl(client), err, httpResp)
	}

	utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
	return nil
}
