package ccpm

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *versionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Custom Connect Plugin Versions.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all versions of a custom connect plugin.",
				Code: "confluent ccpm plugin version list --plugin plugin-123456 --environment env-abcdef",
			},
			examples.Example{
				Text: "List versions of a plugin with output in JSON format.",
				Code: "confluent ccpm plugin version list --plugin plugin-123456 --environment env-abcdef --output json",
			},
		),
	}

	cmd.Flags().String("plugin", "", "Plugin ID.")
	cmd.Flags().String("environment", "", "Environment ID.")
	cobra.CheckErr(cmd.MarkFlagRequired("plugin"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *versionCommand) list(cmd *cobra.Command, args []string) error {
	pluginId, err := cmd.Flags().GetString("plugin")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	versions, err := c.V2Client.ListCCPMPluginVersions(pluginId, environment)
	if err != nil {
		return err
	}

	// Display results in table format
	table := output.NewList(cmd)
	for _, version := range versions {
		spec, _ := version.GetSpecOk()
		status, _ := version.GetStatusOk()

		table.Add(&versionOut{
			Id:                        version.GetId(),
			Version:                   spec.GetVersion(),
			ContentFormat:             spec.GetContentFormat(),
			DocumentationLink:         spec.GetDocumentationLink(),
			SensitiveConfigProperties: spec.GetSensitiveConfigProperties(),
			Phase:                     status.GetPhase(),
			ErrorMessage:              status.GetErrorMessage(),
			Environment:               spec.Environment.Id,
		})
	}

	table.Sort(true)
	return table.Print()
}