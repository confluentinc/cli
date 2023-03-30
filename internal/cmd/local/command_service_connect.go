package local

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/local"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

var connectors = []string{
	"file-sink",
	"file-source",
	"replicator",
}

func NewConnectConnectorCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "connector",
			Short: "Manage connectors.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.AddCommand(NewConnectConnectorConfigCommand(prerunner))
	c.AddCommand(NewConnectConnectorStatusCommand(prerunner))
	c.AddCommand(NewConnectConnectorListCommand(prerunner))
	c.AddCommand(NewConnectConnectorLoadCommand(prerunner))
	c.AddCommand(NewConnectConnectorUnloadCommand(prerunner))

	return c.Command
}

func NewConnectConnectorConfigCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "config <connector-name>",
			Short: "View or set connector configurations.",
			Args:  cobra.ExactArgs(1),
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Print the current configuration of a connector named `s3-sink`:",
					Code: "confluent local services connect connector config s3-sink",
				},
				examples.Example{
					Text: "Configure a connector named `wikipedia-file-source` by passing its configuration properties in JSON format.",
					Code: "confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.json",
				},
				examples.Example{
					Text: "Configure a connector named `wikipedia-file-source` by passing its configuration properties as Java properties.",
					Code: "confluent local services connect connector config wikipedia-file-source --config <path-to-connector>/wikipedia-file-source.properties",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runConnectConnectorConfigCommand
	c.Flags().StringP("config", "c", "", "Configuration file for a connector.")
	return c.Command
}

func (c *Command) runConnectConnectorConfigCommand(cmd *cobra.Command, args []string) error {
	isUp, err := c.isRunning("connect")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("connect")
	}

	connector := args[0]

	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	if configFile == "" {
		out, err := getConnectorConfig(connector)
		if err != nil {
			return err
		}

		output.Printf("Current configuration of %s:\n", connector)
		output.Println(out)
		return nil
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}

	var config map[string]any
	if isJSON(data) {
		if err := json.Unmarshal(data, &config); err != nil {
			return err
		}
		if inner, ok := config["config"]; ok {
			config = inner.(map[string]any)
		}
	} else {
		config = local.ExtractConfig(data)
	}

	config["name"] = connector
	data, err = json.Marshal(config)
	if err != nil {
		return err
	}

	out, err := putConnectorConfig(connector, data)
	if err != nil {
		return err
	}

	output.Println(out)
	return nil
}

func NewConnectConnectorStatusCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "status <connector-name>",
			Short: "Check the status of all connectors, or a single connector.",
			Args:  cobra.MaximumNArgs(1),
		}, prerunner)

	c.Command.RunE = c.runConnectConnectorStatusCommand
	return c.Command
}

func (c *Command) runConnectConnectorStatusCommand(_ *cobra.Command, args []string) error {
	isUp, err := c.isRunning("connect")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("connect")
	}

	if len(args) == 0 {
		out, err := getConnectorsStatus()
		if err != nil {
			return err
		}

		output.Println(out)
		return nil
	}

	connector := args[0]
	out, err := getConnectorStatus(connector)
	if err != nil {
		return err
	}

	output.Println(out)
	return nil
}

func NewConnectConnectorListCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List all bundled connectors.",
			Long:  "List all connectors bundled with Confluent Platform.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.Run = c.runConnectConnectorListCommand
	return c.Command
}

func (c *Command) runConnectConnectorListCommand(_ *cobra.Command, _ []string) {
	output.Println("Bundled Connectors:")
	output.Println(local.BuildTabbedList(connectors))
}

func NewConnectConnectorLoadCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "load <connector-name>",
			Short: "Load a connector.",
			Long:  "Load a bundled connector from Confluent Platform or your own custom connector.",
			Args:  cobra.ExactArgs(1),
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Load a predefined connector called `s3-sink`:",
					Code: "confluent local load s3-sink",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runConnectConnectorLoadCommand
	c.Flags().StringP("config", "c", "", "Configuration file for a connector.")
	return c.Command
}

func (c *Command) runConnectConnectorLoadCommand(cmd *cobra.Command, args []string) error {
	isUp, err := c.isRunning("connect")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("connect")
	}

	connector := args[0]

	var configFile string

	if types.Contains(connectors, connector) {
		configFile, err = c.ch.GetConnectorConfigFile(connector)
		if err != nil {
			return err
		}
	} else {
		configFile, err = cmd.Flags().GetString("config")
		if err != nil {
			return err
		}
		if configFile == "" {
			return fmt.Errorf(errors.InvalidConnectorErrorMsg, connector)
		}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return err
	}
	if !isJSON(data) {
		config := local.ExtractConfig(data)
		delete(config, "name")

		full := map[string]any{
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

	output.Println(out)
	return nil
}

func NewConnectConnectorUnloadCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "unload <connector-name>",
			Short: "Unload a connector.",
			Args:  cobra.ExactArgs(1),
			Example: examples.BuildExampleString(
				examples.Example{
					Text: "Unload a predefined connector called `s3-sink`:",
					Code: "confluent local unload s3-sink",
				},
			),
		}, prerunner)

	c.Command.RunE = c.runConnectConnectorUnloadCommand
	return c.Command
}

func (c *Command) runConnectConnectorUnloadCommand(_ *cobra.Command, args []string) error {
	isUp, err := c.isRunning("connect")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("connect")
	}

	connector := args[0]
	out, err := deleteConnectorConfig(connector)
	if err != nil {
		return err
	}

	if len(out) > 0 {
		output.Println(out)
	} else {
		output.Println("Success.")
	}
	return nil
}

func NewConnectPluginCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "plugin",
			Short: "Manage Connect plugins.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.AddCommand(NewConnectPluginListCommand(prerunner))

	return c.Command
}

func NewConnectPluginListCommand(prerunner cmd.PreRunner) *cobra.Command {
	c := NewLocalCommand(
		&cobra.Command{
			Use:   "list",
			Short: "List available Connect plugins.",
			Long:  "List available Connect plugins bundled with Confluent Platform.",
			Args:  cobra.NoArgs,
		}, prerunner)

	c.Command.RunE = c.runConnectPluginListCommand
	return c.Command
}

func (c *Command) runConnectPluginListCommand(_ *cobra.Command, _ []string) error {
	isUp, err := c.isRunning("connect")
	if err != nil {
		return err
	}
	if !isUp {
		return c.printStatus("connect")
	}

	url := fmt.Sprintf("http://localhost:%d/connector-plugins", services["connect"].port)
	out, err := makeRequest(http.MethodGet, url, []byte{})
	if err != nil {
		return err
	}

	output.Printf(errors.AvailableConnectPluginsMsg, out)
	return nil
}

func isJSON(data []byte) bool {
	var out map[string]any
	return json.Unmarshal(data, &out) == nil
}

func getConnectorConfig(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/config", services["connect"].port, connector)
	return makeRequest(http.MethodGet, url, []byte{})
}

func getConnectorStatus(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/status", services["connect"].port, connector)
	return makeRequest(http.MethodGet, url, []byte{})
}

func getConnectorsStatus() (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors", services["connect"].port)
	return makeRequest(http.MethodGet, url, []byte{})
}

func postConnectorConfig(config []byte) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors", services["connect"].port)
	return makeRequest(http.MethodPost, url, config)
}

func putConnectorConfig(connector string, config []byte) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s/config", services["connect"].port, connector)
	return makeRequest(http.MethodPut, url, config)
}

func deleteConnectorConfig(connector string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/connectors/%s", services["connect"].port, connector)
	return makeRequest(http.MethodDelete, url, []byte{})
}

func makeRequest(method, url string, body []byte) (string, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := utils.DefaultClient()
	client.Timeout = 10 * time.Second
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}

	return formatJSONResponse(res)
}

func formatJSONResponse(res *http.Response) (string, error) {
	out, err := io.ReadAll(res.Body)
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
