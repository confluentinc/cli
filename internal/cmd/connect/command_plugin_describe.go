package connect

import (
	"fmt"

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

	reply, _, err := c.V2Client.ValidateConnectorPlugin(args[0], c.EnvironmentId(), kafkaCluster.ID, config)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.InvalidCloudErrorMsg, errors.InvalidCloudSuggestions)
	}

	outputFormat, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		return err
	}

	if outputFormat == output.Human.String() {
		utils.Println(cmd, "The following are required configs:")
		utils.Println(cmd, "connector.class : "+args[0])
		for _, c := range *reply.Configs {
			if len(c.Value.GetErrors()) > 0 {
				utils.Println(cmd, c.Value.GetName()+" : ["+c.Value.GetErrors()[0]+"]")
			}
		}
		return nil
	}

	for _, c := range *reply.Configs {
		if len(c.Value.GetErrors()) > 0 {
			config[c.Value.GetName()] = fmt.Sprintf("%s ", c.Value.GetErrors()[0])
		}
	}
	return output.StructuredOutput(outputFormat, &config)
}
