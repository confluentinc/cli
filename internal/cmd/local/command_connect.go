package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

var connectorConfigs = map[string]string{
	"elasticsearch-sink": "kafka-connect-elasticsearch/quickstart-elasticsearch.properties",
	"file-sink":          "kafka/connect-file-sink.properties",
	"file-source":        "kafka/connect-file-source.properties",
	"hdfs-sink":          "kafka-connect-hdfs/quickstart-hdfs.properties",
	"jdbc-sink":          "kafka-connect-jdbc/sink-quickstart-sqlite.properties",
	"jdbc-source":        "kafka-connect-jdbc/source-quickstart-sqlite.properties",
	"s3-sink":            "kafka-connect-s3/quickstart-s3.properties",
}

func NewConnectConnectorCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "connector",
			Short: "Manage connectors.",
			Args:  cobra.NoArgs,
		}, cfg, prerunner)

	connectConnectorCommand.AddCommand(NewConnectConnectorConfigCommand(prerunner, cfg))
	connectConnectorCommand.AddCommand(NewConnectConnectorStatusCommand(prerunner, cfg))
	connectConnectorCommand.AddCommand(NewConnectConnectorListCommand(prerunner, cfg))
	connectConnectorCommand.AddCommand(NewConnectConnectorLoadCommand(prerunner, cfg))
	connectConnectorCommand.AddCommand(NewConnectConnectorUnloadCommand(prerunner, cfg))

	return connectConnectorCommand.Command
}

func NewConnectConnectorConfigCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorConfigCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "config [connector]",
			Short: "Print a connector config, or configure an existing connector.",
			Args:  cobra.ExactArgs(1),
			RunE:  runConnectConnectorConfigCommand,
		}, cfg, prerunner)

	connectConnectorConfigCommand.Flags().StringP("config", "c", "", "Configuration file for a connector.")

	return connectConnectorConfigCommand.Command
}

func runConnectConnectorConfigCommand(command *cobra.Command, args []string) error {
	connector := args[0]

	configFile, err := command.Flags().GetString("config")
	if err != nil {
		return err
	}
	if configFile == "" {
		out, err := getConnectorConfig(connector)
		if err != nil {
			return err
		}

		command.Printf("Current configuration of %s:\n", connector)
		command.Println(out)
		return nil
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	if !isJSON(data) {
		config := extractConfig(data)
		data, err = json.Marshal(config)
		if err != nil {
			return err
		}
	}

	out, err := putConnectorConfig(connector, data)
	if err != nil {
		return err
	}

	command.Println(out)
	return nil
}

func NewConnectConnectorStatusCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorStatusCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "status [connector]",
			Short: "Check the status of all connectors, or a single connector.",
			Args:  cobra.MaximumNArgs(1),
			RunE:  runConnectConnectorStatusCommand,
		}, cfg, prerunner)

	return connectConnectorStatusCommand.Command
}

func runConnectConnectorStatusCommand(command *cobra.Command, args []string) error {
	if len(args) == 0 {
		out, err := getConnectorsStatus()
		if err != nil {
			return err
		}

		command.Println(out)
		return nil
	}

	connector := args[0]
	out, err := getConnectorStatus(connector)
	if err != nil {
		return err
	}

	command.Println(out)
	return nil
}

func NewConnectConnectorListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List connectors.",
			Args:  cobra.NoArgs,
			Run:   runConnectConnectorListCommand,
		},
		cfg, prerunner)

	return connectConnectorListCommand.Command
}

func runConnectConnectorListCommand(command *cobra.Command, _ []string) {
	command.Println("Bundled Predefined Connectors:")
	command.Println(buildTabbedList(getConnectors()))
}

func NewConnectConnectorLoadCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorLoadCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "load [connector]",
			Short: "Load a connector.",
			Args:  cobra.ExactArgs(1),
			RunE:  runConnectConnectorLoadCommand,
		},
		cfg, prerunner)

	connectConnectorLoadCommand.Flags().StringP("config", "c", "", "Configuration file for a connector.")

	return connectConnectorLoadCommand.Command
}

func runConnectConnectorLoadCommand(command *cobra.Command, args []string) error {
	connector := args[0]

	configFile, ok := connectorConfigs[connector]
	if ok {
		confluentHome, err := getConfluentHome()
		if err != nil {
			return err
		}
		configFile = filepath.Join(confluentHome, "etc", configFile)
	} else {
		file, err := command.Flags().GetString("config")
		if err != nil {
			return err
		}
		if file == "" {
			return fmt.Errorf("invalid connector: %s", connector)
		}
		configFile = file
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	if !isJSON(data) {
		config := extractConfig(data)
		delete(config, "name")

		full := map[string]interface{}{
			"name":   connector,
			"config": config,
		}

		data, err = json.Marshal(full)
		if err != nil {
			return err
		}
	}

	out, err := postConnectorConfig(data)
	if err != nil {
		return err
	}

	command.Println(out)
	return nil
}

func NewConnectConnectorUnloadCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectConnectorUnloadCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "unload [connector]",
			Short: "Unload a connector.",
			Args:  cobra.ExactArgs(1),
			RunE:  runConnectConnectorUnloadCommand,
		}, cfg, prerunner)

	return connectConnectorUnloadCommand.Command
}

func runConnectConnectorUnloadCommand(command *cobra.Command, args []string) error {
	connector := args[0]
	out, err := deleteConnectorConfig(connector)
	if err != nil {
		return err
	}

	if len(out) > 0 {
		command.Println(out)
	} else {
		command.Println("Success.")
	}
	return nil
}

func NewConnectPluginCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectPluginCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "plugin",
			Short: "Manage connect plugins.",
			Args:  cobra.NoArgs,
		},
		cfg, prerunner)

	connectPluginCommand.AddCommand(NewConnectPluginListCommand(prerunner, cfg))

	return connectPluginCommand.Command
}

func NewConnectPluginListCommand(prerunner cmd.PreRunner, cfg *v3.Config) *cobra.Command {
	connectPluginListCommand := cmd.NewAnonymousCLICommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available connect plugins.",
			Args:  cobra.NoArgs,
			RunE:  runConnectPluginListCommand,
		},
		cfg, prerunner)

	return connectPluginListCommand.Command
}

func runConnectPluginListCommand(command *cobra.Command, _ []string) error {
	url := fmt.Sprintf("http://localhost:%d/connector-plugins", services["connect"].port)
	out, err := makeRequest("GET", url, []byte{})
	if err != nil {
		return err
	}

	command.Println("Available Connect Plugins:")
	command.Println(out)
	return nil
}

func getConnectors() []string {
	var connectors []string
	for connector := range connectorConfigs {
		connectors = append(connectors, connector)
	}
	return connectors
}

func isJSON(data []byte) bool {
	var out map[string]interface{}
	return json.Unmarshal(data, &out) == nil
}

func extractConfig(data []byte) map[string]string {
	re := regexp.MustCompile(`(?m)^[^\s#]*=.+`)
	matches := re.FindAllString(string(data), -1)
	config := map[string]string{}

	for _, match := range matches {
		x := strings.Split(match, "=")
		key, val := x[0], x[1]
		config[key] = val
	}
	return config
}

func getConnectorConfig(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/config", services["connect"].port, connector)
	return makeRequest("GET", url, []byte{})
}

func getConnectorStatus(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/status", services["connect"].port, connector)
	return makeRequest("GET", url, []byte{})
}

func getConnectorsStatus() (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors", services["connect"].port)
	return makeRequest("GET", url, []byte{})
}

func postConnectorConfig(config []byte) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors", services["connect"].port)
	return makeRequest("POST", url, config)
}

func putConnectorConfig(connector string, config []byte) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/config", services["connect"].port, connector)
	return makeRequest("PUT", url, config)
}

func deleteConnectorConfig(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s", services["connect"].port, connector)
	return makeRequest("DELETE", url, []byte{})
}

func makeRequest(method, url string, body []byte) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("start the connect service with \"confluent local services connect start\"")
	}

	return formatJSONResponse(res)
}

func formatJSONResponse(res *http.Response) (string, error) {
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	buf := new(bytes.Buffer)
	if len(out) > 0 {
		err = json.Indent(buf, out, "", "  ")
		if err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}
