package ccpm

import (
	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/spf13/cobra"
)

func (c *pluginCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Custom Connect Plugin.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	cmd.Flags().String("name", "", "Display name of the Custom Connect Plugin.")
	cmd.Flags().String("description", "", "Description of the Custom Connect Plugin.")
	cmd.Flags().String("cloud", "", "Cloud provider (AWS, GCP, AZURE).")
	cmd.Flags().String("environment", "", "Environment ID.")
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
