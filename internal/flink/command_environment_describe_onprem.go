package flink

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newEnvironmentDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink Environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.describeFlinkEnvironment,
	}

	return cmd
}

func (c *command) describeFlinkEnvironment(cmd *cobra.Command, args []string) error {
	cmfRest, err := c.GetCmfREST()
	if err != nil {
		return err
	}

	// Get the name of the application to be retrieved
	environmentName := args[0]
	cmfEnvironment, httpResponse, err := cmfRest.Client.DefaultApi.GetEnvironment(cmd.Context(), environmentName)

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

	// In case err != nil but status code is 200 - if that's possible.
	if err != nil {
		return fmt.Errorf("failed to describe environment \"%s\": %s", environmentName, err)
	}

	if output.GetFormat(cmd) == output.Human {
		environmentTable := output.NewTable(cmd)
		environmentTable.Add(&flinkEnvironmentOut{
			Name:            cmfEnvironment.Name,
			DefaultStrategy: cmfEnvironment.DefaultStrategy,
			CreatedTime:     cmfEnvironment.CreatedTime.String(),
			UpdatedTime:     cmfEnvironment.UpdatedTime.String(),
		})
		return environmentTable.Print()
	}
	return output.SerializedOutput(cmd, cmfEnvironment)

}
