package flink

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *unauthenticatedCommand) newEnvironmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink Environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentDescribe,
	}

	return cmd
}

func (c *unauthenticatedCommand) environmentDescribe(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the environment to be retrieved
	environmentName := args[0]
	cmfEnvironment, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), environmentName)

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
			return fmt.Errorf("environment \"%s\" not found %s", environmentName, string(respBody))
		case http.StatusInternalServerError:
			return fmt.Errorf("internal server error while describing environment \"%s\": %s", environmentName, string(respBody))
		default:
			return fmt.Errorf("failed to describe environment \"%s\": %s", environmentName, err)
		}
	}

	table := output.NewTable(cmd)
	var defaultsBytes []byte
	defaultsBytes, err = json.Marshal(cmfEnvironment.Defaults)

	table.Add(&flinkEnvironmentOutput{
		Name:        cmfEnvironment.Name,
		Defaults:    string(defaultsBytes),
		CreatedTime: cmfEnvironment.CreatedTime.String(),
		UpdatedTime: cmfEnvironment.UpdatedTime.String(),
	})
	return table.Print()

}
