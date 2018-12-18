package common

import (
	"fmt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"os"
	"os/exec"

	"github.com/codyaray/go-editor"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/shared"
)

var messages = map[error]string{
	shared.ErrNoContext:      "You must login to access Confluent Cloud.",
	shared.ErrUnauthorized:   "You must login to access Confluent Cloud.",
	shared.ErrExpiredToken:   "Your access to Confluent Cloud has expired. Please login again.",
	shared.ErrIncorrectAuth:  "You have entered an incorrect username or password. Please try again.",
	shared.ErrMalformedToken: "Your auth token has been corrupted. Please login again.",
	shared.ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	shared.ErrNotFound:       "Kafka cluster not found.", // TODO: parametrize ErrNotFound for better error messaging
}

// HandleError provides standard error messaging for common errors.
func HandleError(err error, cmd *cobra.Command) error {
	out := cmd.OutOrStderr()
	if msg, ok := messages[err]; ok {
		fmt.Fprintln(out, msg)
		return nil
	}

	switch err.(type) {
	case editor.ErrEditing:
		fmt.Fprintln(out, err)
	case shared.NotAuthenticatedError:
		fmt.Fprintln(out, err)
	case shared.KafkaError:
		fmt.Fprintln(out, err)
	default:
		return err
	}
	return nil
}

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
