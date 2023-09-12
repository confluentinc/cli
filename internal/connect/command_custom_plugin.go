package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	v1 "github.com/confluentinc/cli/v3/pkg/config"
)

type customPluginCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type customPluginOut struct {
	Id                  string   `human:"ID" serialized:"id"`
	Name                string   `human:"Name" serialized:"name"`
	Description         string   `human:"Description" serialized:"description"`
	ConnectorClass      string   `human:"Connector Class" serialized:"connector_class"`
	ConnectorType       string   `human:"Connector Type" serialized:"connector_type"`
	SensitiveProperties []string `human:"Sensitive Properties" serialized:"sensitive_properties"`
}

func newCustomPluginCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "custom-plugin",
		Short:       "Manage custom plugins.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := new(customPluginCommand)

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newListCommand())
		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newUpdateCommand())
	}

	return cmd
}
