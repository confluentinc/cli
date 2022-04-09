package kafka

import (
	cloudkafkarest "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/properties"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *linkCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <link>",
		Short: "Update link configs.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update configuration values for the cluster link `my-link`.",
				Code: "confluent kafka link update my-link --config-file my-config.txt",
			},
		),
	}

	cmd.Flags().String(configFileFlagName, "", "Name of the file containing link config overrides. "+
		"Each property key-value pair should have the format of key=value. Properties are separated by new-line characters.")

	if c.cfg.IsCloudLogin() {
		pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = cmd.MarkFlagRequired(configFileFlagName)

	return cmd
}

func (c *linkCommand) update(cmd *cobra.Command, args []string) error {
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

	kafkaREST, err := c.GetCloudKafkaREST()
	if err != nil {
		return err
	}

	kafkaClusterConfig, err := c.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}
	clusterId := kafkaClusterConfig.ID

	req := kafkaREST.Client.ClusterLinkingV3Api.UpdateKafkaLinkConfigBatch(kafkaREST.Context, clusterId, linkName)
	httpResp, err := req.AlterConfigBatchRequestData(cloudkafkarest.AlterConfigBatchRequestData{Data: toCloudAlterConfigBatchRequestData(configMap)}).Execute()
	if err != nil {
		return kafkaRestError(pcmd.GetCloudKafkaRestBaseUrl(kafkaREST.Client), err, httpResp)
	}

	utils.Printf(cmd, errors.UpdatedLinkMsg, linkName)
	return nil
}
