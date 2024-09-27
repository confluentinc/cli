package flink

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func (c *unauthenticatedCommand) newEnvironmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentCreate,
	}

	cmd.Flags().String("defaults", "", "JSON string defining the environment defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension)")
	return cmd
}

func (c *unauthenticatedCommand) environmentCreate(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environmentName := args[0]

	// Read file contents or parse defaults if applicable
	var defaultsParsed map[string]interface{}
	defaults, err := cmd.Flags().GetString("defaults")
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	if defaults != "" {
		defaultsParsed = make(map[string]interface{})
		if strings.HasSuffix(defaults, ".json") {
			data, err := ioutil.ReadFile(defaults)
			if err != nil {
				return fmt.Errorf("failed to read defaults file: %v", err)
			}
			err = json.Unmarshal(data, &defaultsParsed)
		} else if strings.HasSuffix(defaults, ".yaml") || strings.HasSuffix(defaults, ".yml") {
			data, err := ioutil.ReadFile(defaults)
			if err != nil {
				return fmt.Errorf("failed to read defaults file: %v", err)
			}
			err = yaml.Unmarshal(data, &defaultsParsed)
		} else {
			err = json.Unmarshal([]byte(defaults), &defaultsParsed)
		}

		if err != nil {
			return fmt.Errorf("failed to parse defaults: %v", err)
		}
	}

	_, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), environmentName)
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == 200 {
		return fmt.Errorf("environment \"%s\" already exists", environmentName)
	}

	var environment cmfsdk.PostEnvironment
	environment.Name = environmentName
	if defaultsParsed != nil {
		environment.Defaults = defaultsParsed
	}
	outputEnvironment, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateEnvironment(cmd.Context(), environment)
	defer httpResponse.Body.Close()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 201 {
			if httpResponse.Body != nil {
				respBody, parseError := ioutil.ReadAll(httpResponse.Body)
				if parseError == nil {
					return fmt.Errorf("failed to create environment \"%s\": %s", environmentName, string(respBody))
				}
			}
		}
		return fmt.Errorf("failed to create environment \"%s\": %s", environmentName, err)
	}
	// TODO: can err == nil and status code non-20x?

	if output.GetFormat(cmd) == output.Human {
		// TODO: Add different output formats
	}
	return output.SerializedOutput(cmd, outputEnvironment)
}
