package flink

import (
	"fmt"
	"io"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

func (c *command) newApplicationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Applications.",
		RunE:  c.listApplicationsOnPrem,
	}

	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *command) listApplicationsOnPrem(cmd *cobra.Command, _ []string) error {
	environmentName, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentName == "" {
		fmt.Errorf("Environment name is required")
		return nil
	}

	cmfREST, err := c.GetCmfREST()
	if err != nil {
		return err
	}

	applicationsPage, httpResponse, err := cmfREST.Client.DefaultApi.GetApplications(cmd.Context(), environmentName, nil)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			if httpResponse.Body != nil {
				defer httpResponse.Body.Close()
				respBody, parseError := io.ReadAll(httpResponse.Body)
				if parseError == nil {
					return fmt.Errorf("failed to list applications in the environment \"%s\": %s", environmentName, string(respBody))
				}
			}
		}
		return err
	}

	var list []cmfsdk.Application
	applications := append(list, applicationsPage.Items...)

	// TODO: Add pagination support once the API supports it
	if len(applications) == 0 {
		return fmt.Errorf("no applications found in the environment \"%s\"", environmentName)
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			jobStatus := app.Status["jobStatus"].(map[string]interface{})
			environment, ok := app.Metadata["environment"].(string)
			if !ok {
				environment = environmentName
			}
			list.Add(&flinkApplicationOut{
				Name:        app.Metadata["name"].(string),
				Environment: environment,
				JobId:       jobStatus["jobId"].(string),
				JobState:    jobStatus["state"].(string),
			})
		}
		return list.Print()
	}
	return output.SerializedOutput(cmd, applications)
}
