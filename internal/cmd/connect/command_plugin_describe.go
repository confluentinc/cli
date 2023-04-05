package connect

import (
	"sort"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type pluginDescribeOut struct {
	Config        string `human:"Config" serialized:"config"`
	Documentation string `human:"Documentation" serialized:"documentation"`
	IsRequired    bool   `human:"Required" serialized:"is_required"`
}

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
				Code: "confluent connect plugin describe MySource",
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

	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	reply, err := c.V2Client.ValidateConnectorPlugin(args[0], environmentId, kafkaCluster.ID, config)
	if err != nil {
		return errors.NewWrapErrorWithSuggestions(err, errors.InvalidCloudErrorMsg, errors.InvalidCloudSuggestions)
	}

	list := output.NewList(cmd)
	list.Sort(false)

	configs := reply.GetConfigs()
	sort.Slice(configs, func(i, j int) bool {
		requiredI := configs[i].Definition.GetRequired()
		requiredJ := configs[j].Definition.GetRequired()
		if requiredI == requiredJ {
			return configs[i].Value.GetName() < configs[j].Value.GetName()
		}

		return requiredI
	})

	for _, config := range configs {
		doc := config.Definition.GetDisplayName()
		if config.Definition.GetDocumentation() != "" {
			doc = config.Definition.GetDocumentation()
		}
		list.Add(&pluginDescribeOut{
			Config:        config.Value.GetName(),
			Documentation: doc,
			IsRequired:    config.Definition.GetRequired(),
		})
	}

	return list.Print()
}
