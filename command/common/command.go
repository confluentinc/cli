package common

import (
	"fmt"
	"github.com/confluentinc/cli/shared"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
)

type CommandFactory func(config *shared.Config, plugin interface{}) *cobra.Command

func InitCommand(pluginPath string, pluginName string, config *shared.Config, command *cobra.Command, commandFactories []CommandFactory) error {
	path, err := exec.LookPath(pluginPath)
	if err != nil {
		return fmt.Errorf("skipping %s: plugin isn't installed", pluginName)
	}

	// We're a host. Start by launching the plugin process.
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  shared.Handshake,
		Plugins:          shared.PluginMap,
		Cmd:              exec.Command("sh", "-c", path), // nolint: gas
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Managed:          true,
		Logger: hclog.New(&hclog.LoggerOptions{
			Output: hclog.DefaultOutput,
			Level:  hclog.Info,
			Name:   "plugin",
		}),
	})

	// Connect via RPC.
	rpcClient, err := client.Client()
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(pluginName)
	if err != nil {
		fmt.Println("Error:", err.Error())
		os.Exit(1)
	}

	// All commands require login first
	command.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if err = config.CheckLogin(); err != nil {
			_ = HandleError(err, cmd)
			os.Exit(1)
		}
	}

	for _, factory := range commandFactories {
		command.AddCommand(factory(config, raw))
	}
	return nil
}
