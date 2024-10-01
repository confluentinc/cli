package flink

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newEnvironmentCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentCreate,
	}

	cmd.Flags().String("defaults", "", "JSON string defining the environment defaults, or path to a file to read defaults from (with .yml, .yaml or .json extension)")
	return cmd
}

func (c *command) environmentCreate(cmd *cobra.Command, args []string) error {
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
			var data []byte
			data, err = os.ReadFile(defaults)
			if err != nil {
				return fmt.Errorf("failed to read defaults file: %v", err)
			}
			err = json.Unmarshal(data, &defaultsParsed)
		} else if strings.HasSuffix(defaults, ".yaml") || strings.HasSuffix(defaults, ".yml") {
			var data []byte
			data, err = os.ReadFile(defaults)
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

	_, httpResponse, _ := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), environmentName)
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusOK {
		return fmt.Errorf(`environment "%s" already exists`, environmentName)
	}

	var environment cmfsdk.PostEnvironment
	environment.Name = environmentName
	if defaultsParsed != nil {
		environment.Defaults = defaultsParsed
	}
	outputEnvironment, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateEnvironment(cmd.Context(), environment)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to create environment "%s": %s`, environmentName, parsedErr)
	}

	table := output.NewTable(cmd)
	var defaultsBytes []byte
	defaultsBytes, err = json.Marshal(outputEnvironment.Defaults)
	if err != nil {
		return fmt.Errorf("failed to marshal defaults: %s", err)
	}

	table.Add(&flinkEnvironmentOutput{
		Name:        outputEnvironment.Name,
		Defaults:    string(defaultsBytes),
		CreatedTime: outputEnvironment.CreatedTime.String(),
		UpdatedTime: outputEnvironment.UpdatedTime.String(),
	})
	return table.Print()
}
