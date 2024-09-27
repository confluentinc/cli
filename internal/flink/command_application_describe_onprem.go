package flink

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *unauthenticatedCommand) newApplicationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a flinkApplication.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationDescribe,
	}

	cmd.Flags().String("environment", "", "REQUIRED: Name of the Environment for the Flink Application.")
	cmd.MarkFlagRequired("environment")

	return cmd
}

func (c *unauthenticatedCommand) applicationDescribe(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environment == "" {
		return errors.New("environment is required")
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the application to be retrieved
	applicationName := args[0]
	cmfApplication, httpResponse, err := cmfClient.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)

	if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
		// Read response body if any
		respBody := []byte{}
		var parseError error
		if httpResponse.Body != nil {
			defer httpResponse.Body.Close()
			respBody, parseError = ioutil.ReadAll(httpResponse.Body)
			if parseError != nil {
				respBody = []byte(fmt.Sprintf("failed to read response body: %s", parseError))
			}
		}
		// Start checking the possible status codes
		switch httpResponse.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("application \"%s\" not found %s", applicationName, string(respBody))
		case http.StatusInternalServerError:
			return fmt.Errorf("internal server error while describing application \"%s\": %s", applicationName, string(respBody))
		default:
			return fmt.Errorf("failed to describe application \"%s\": %s", applicationName, err)
		}
	}

	// In case err != nil but status code is 200 - if that's possible.
	if err != nil {
		return fmt.Errorf("failed to describe application \"%s\" in the environment \"%s\": %s", applicationName, environment, err)
	}

	if output.GetFormat(cmd) == output.Human {
		applicationTable := output.NewTable(cmd)
		jobStatus := cmfApplication.Status["jobStatus"].(map[string]interface{})
		applicationTable.Add(&flinkApplicationSummary{
			Name:        cmfApplication.Metadata["name"].(string),
			Environment: environment,
			JobId:       jobStatus["jobId"].(string),
			JobState:    jobStatus["state"].(string),
		})
		return applicationTable.Print()
	}
	return output.SerializedOutput(cmd, cmfApplication)

}
