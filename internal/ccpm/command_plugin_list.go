package ccpm

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/output"
)

type pluginOut struct {
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Description string `human:"Description" serialized:"description"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Language    string `human:"Runtime Language" serialized:"runtime_language"`
}

func (c *pluginCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Custom Connect Plugins.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().String("cloud", "", "Filter by cloud provider (AWS, GCP, AZURE).")
	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.MarkFlagRequired("environment")

	return cmd
}

func (c *pluginCommand) list(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Use V2Client to call CCPM API
	pluginList, err := c.V2Client.ListCCPMPlugins(cloud, environment)
	if err != nil {
		return err
	}

	// Display results in table format
	table := output.NewTable(cmd)
	for _, plugin := range pluginList.GetData() {
		spec, _ := plugin.GetSpecOk()
		table.Add(&pluginOut{
			Id:          plugin.GetId(),
			Name:        spec.GetDisplayName(),
			Description: spec.GetDescription(),
			Cloud:       spec.GetCloud(),
			Language:    spec.GetRuntimeLanguage(),
		})
	}

	return table.Print()
}
