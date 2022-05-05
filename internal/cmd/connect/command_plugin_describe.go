package connect

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *pluginCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <plugin>",
		Short:             "Describe a connector plugin.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the required connector configuration parameters for connector plugin "MySource".`,
				Code: fmt.Sprintf("%s connect plugin describe MySource", version.CLIName),
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *pluginCommand) describe(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	config := map[string]string{"connector.class": args[0]}
	connectorConfig := &schedv1.ConnectorConfig{
		UserConfigs:    config,
		AccountId:      c.EnvironmentId(),
		KafkaClusterId: kafkaCluster.ID,
		Plugin:         args[0],
	}

	reply, err := c.Client.Connect.Validate(context.Background(), connectorConfig)
	if reply != nil && err != nil {
		outputFormat, flagErr := cmd.Flags().GetString(output.FlagName)
		if flagErr != nil {
			return flagErr
		}

		if outputFormat == output.Human.String() {
			utils.Println(cmd, "The following are required configs:")
			utils.Print(cmd, "connector.class : "+args[0]+"\n"+err.Error())
			return nil
		}

		for _, c := range reply.Configs {
			if len(c.Value.Errors) > 0 {
				config[c.Value.Name] = fmt.Sprintf("%s ", c.Value.Errors[0])
			}
		}
		return output.StructuredOutput(outputFormat, &config)
	}

	return errors.Errorf(errors.InvalidCloudErrorMsg)
}
