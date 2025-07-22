package ccpm

import (
	"github.com/spf13/cobra"

	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type pluginCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type pluginOut struct {
	Id              string `human:"ID" serialized:"id"`
	Name            string `human:"Name" serialized:"name"`
	Description     string `human:"Description" serialized:"description"`
	Cloud           string `human:"Cloud" serialized:"cloud"`
	RuntimeLanguage string `human:"Runtime Language" serialized:"runtime_language"`
	Environment     string `human:"Environment" serialized:"environment"`
}

func newPluginCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "plugin",
		Short:       "Manage custom Connect plugins.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &pluginCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newVersionCommand())

	return cmd
}

func printCustomConnectPluginTable(cmd *cobra.Command, plugin ccpmv1.CcpmV1CustomConnectPlugin) error {
	table := output.NewTable(cmd)

	table.Add(&pluginOut{
		Id:              plugin.GetId(),
		Name:            plugin.Spec.GetDisplayName(),
		Description:     plugin.Spec.GetDescription(),
		Cloud:           plugin.Spec.GetCloud(),
		RuntimeLanguage: plugin.Spec.GetRuntimeLanguage(),
		Environment:     plugin.GetSpec().Environment.GetId(),
	})

	return table.Print()
}
