package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v3"
)

var (
	services = []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
	}
	confluentPlatformServices = []string{
		"control-center",
	}

	connectors = []string{
		"elasticsearch-sink",
		"file-source",
		"file-sink",
		"jdbc-source",
		"jdbc-sink",
		"hdfs-sink",
		"s3-sink",
	}
)

func NewListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	listCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list [command]",
			Short: "List all the available services, connectors, or plugins.",
			Args:  cobra.MaximumNArgs(1),
			RunE:  runListCommand,
		},
		cfg, prerunner)

	listCommand.AddCommand(NewListConnectorsCommand(prerunner, cfg))
	listCommand.AddCommand(NewListPluginsCommand(prerunner, cfg))

	return listCommand.Command
}

func runListCommand(command *cobra.Command, _ []string) error {
	isCP, err := isConfluentPlatform()
	if err != nil {
		return err
	}

	availableServices := services
	if isCP {
		availableServices = append(services, confluentPlatformServices...)
	}

	command.Println("Available Services:")
	command.Println(buildTabbedList(availableServices))

	return nil
}

func NewListConnectorsCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectorsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "connectors",
			Short: "List all the available connectors.",
			Args:  cobra.NoArgs,
			RunE:  runListConnectorsCommand,
		},
		cfg, prerunner)

	return connectorsCommand.Command
}

func runListConnectorsCommand(command *cobra.Command, _ []string) error {
	command.Println("Bundled Predefined Connectors:")
	command.Println(buildTabbedList(connectors))

	return nil
}

func NewListPluginsCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	pluginsCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "plugins",
			Short: "List all the available plugins.",
			Args:  cobra.NoArgs,
			RunE:  runListPluginsCommand,
		},
		cfg, prerunner)

	return pluginsCommand.Command
}

func runListPluginsCommand(command *cobra.Command, _ []string) error {
	port := 8083
	url := fmt.Sprintf("http://localhost:%d/connector-plugins", port)

	plugins, err := dumpJSON(url)
	if err != nil {
		return err
	}

	command.Println("Available Connector Plugins:")
	command.Println(plugins)

	return nil
}

func buildTabbedList(slice []string) string {
	var list strings.Builder
	for _, x := range slice {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
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
