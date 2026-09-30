package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
)

type customRuntimeCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type customRuntimeOut struct {
	Id                             string   `human:"ID" serialized:"id"`
	CustomConnectPluginRuntimeName string   `human:"Runtime Name" serialized:"custom_connect_plugin_runtime_name"`
	RuntimeAkVersion               string   `human:"Runtime AK Version" serialized:"runtime_ak_version"`
	SupportedJavaVersions          []string `human:"Supported Java Versions" serialized:"supported_java_versions"`
	ProductMaturity                string   `human:"Product Maturity" serialized:"product_maturity"`
	EndOfLifeAt                    string   `human:"End of Life At" serialized:"end_of_life_at"`
	Description                    string   `human:"Description" serialized:"description"`
}

func newCustomRuntimeCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "custom-connector-runtime",
		Short:       "Manage custom connector runtimes.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &customRuntimeCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newListCommand())

	_ = cfg.ParseFlagsIntoConfig(cmd)

	return cmd
}
