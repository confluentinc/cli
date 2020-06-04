package local

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

func NewPluginsCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	pluginsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "plugins [command]",
			Short: "Manage plugins for Kafka Connect.",
			Args:  cobra.ExactArgs(1),
		},
		cfg, prerunner)

	pluginsCommand.AddCommand(NewPluginsListCommand(prerunner, cfg))

	return pluginsCommand.Command
}

func NewPluginsListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	pluginsListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available plugins for Kafka Connect.",
			Args:  cobra.NoArgs,
			RunE:  runListPluginsCommand,
		},
		cfg, prerunner)

	return pluginsListCommand.Command
}

func runListPluginsCommand(command *cobra.Command, _ []string) error {
	plugins, err := dumpJSON("http://localhost:8083/connector-plugins")
	if err != nil {
		return err
	}

	command.Println("Available Connector Plugins:")
	command.Println(plugins)

	return nil
}

func dumpJSON(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}

	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	err = json.Indent(buf, out, "", "  ")
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}
