package ccpm

import (
	ccpmv1 "github.com/confluentinc/ccloud-sdk-go-v2/ccpm/v1"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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
		Use:   "plugin",
		Short: "Manage Custom Connect Plugins.",
	}

	c := &pluginCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())
	cmd.AddCommand(c.newVersionCommand())

	return cmd
}

func printCustomConnectPluginTable(cmd *cobra.Command, plugin ccpmv1.CcpmV1CustomConnectPlugin) error {
	table := output.NewTable(cmd)

	table.Add(&pluginOut{
		Id:              *plugin.Id,
		Name:            *plugin.Spec.DisplayName,
		Description:     *plugin.Spec.Description,
		Cloud:           *plugin.Spec.Cloud,
		RuntimeLanguage: *plugin.Spec.RuntimeLanguage,
		Environment:     plugin.Spec.Environment.Id,
	})

	return table.Print()
}
