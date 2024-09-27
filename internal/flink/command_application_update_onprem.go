package flink

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func (c *command) newApplicationUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.updateApplicationOnPrem,
	}

	return cmd
}

func (c *command) updateApplicationOnPrem(cmd *cobra.Command, args []string) error {
	environment := getEnvironment(cmd)
	if environment == "" {
		return errors.New("environment name is required. You can use the --environment flag or set the default environment using `confluent flink environment use <name>` command")
	}

	cmfREST, err := c.GetCmfREST()
	if err != nil {
		return err
	}

	// Check if the application already exists
	resourceFilePath := args[0]
	// Read file contents
	data, err := ioutil.ReadFile(resourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	var application cmfsdk.Application
	ext := filepath.Ext(resourceFilePath)
	switch ext {
	case ".json":
		err = json.Unmarshal(data, &application)
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &application)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
	if err != nil {
		return err
	}

	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, err := cmfREST.Client.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode != 200 {
		return fmt.Errorf("application \"%s\" does not exist in the environment \"%s\"", applicationName, environment)
	}

	outputApplication, httpResponse, err := cmfREST.Client.DefaultApi.CreateOrUpdateApplication(cmd.Context(), environment, application)
	defer httpResponse.Body.Close()
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			respBody, parseError := ioutil.ReadAll(httpResponse.Body)
			if httpResponse.Body != nil {
				if parseError == nil {
					return fmt.Errorf("failed to update application \"%s\" in the environment \"%s\": %s", applicationName, environment, string(respBody))
				}
			}
		}
		return fmt.Errorf("failed to update application \"%s\" in the environment \"%s\": %s", applicationName, environment, err)
	}
	// TODO: can err == nil and status code non-20x?

	if output.GetFormat(cmd) == output.Human {
		// TODO: Add different output formats
	}
	return output.SerializedOutput(cmd, outputApplication)
}
