package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *pluginCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom Connect plugin.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a custom Connect plugin for AWS.",
				Code: "confluent ccpm plugin create --name \"My Custom Plugin\" --description \"A custom connector for data processing\" --cloud AWS --environment env-12345",
			},
			examples.Example{
				Text: "Create a custom Connect plugin for GCP with minimal description.",
				Code: "confluent ccpm plugin create --name \"GCP Data Connector\" --cloud GCP --environment env-abcdef",
			},
		),
	}

	cmd.Flags().String("environment", "", "Environment ID.")
	cmd.Flags().String("cloud", "", "Cloud provider (AWS, GCP, AZURE).")
	cmd.Flags().String("name", "", "Display name of the custom Connect plugin.")
	cmd.Flags().String("description", "", "Description of the custom Connect plugin.")
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("name"))
	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *pluginCommand) create(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	description, err := cmd.Flags().GetString("description")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Create CCPM plugin request
	request := ccpmv1.CcpmV1CustomConnectPlugin{
		Spec: &ccpmv1.CcpmV1CustomConnectPluginSpec{
			DisplayName: &name,
			Description: &description,
			Cloud:       &cloud,
			Environment: &ccpmv1.EnvScopedObjectReference{Id: environment},
		},
	}

	// Use V2Client to call CCPM API
	plugin, err := c.V2Client.CreateCCPMPlugin(request)
	if err != nil {
		return err
	}
	return printCustomConnectPluginTable(cmd, plugin)
}
