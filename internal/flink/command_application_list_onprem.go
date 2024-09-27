package flink

import (
	"fmt"
	"io"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

func (c *command) newApplicationListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().StringP("environment", "e", "", "REQUIRED: Name of the Environment to get the FlinkApplication from.")
	cmd.MarkFlagRequired("environment")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) applicationList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environment == "" {
		log.CliLogger.Error("environment is required")
		return nil
	}

	cmfREST, err := c.GetCmfRest()
	if err != nil {
		return err
	}

	applicationsPage, httpResponse, err := cmfREST.Client.DefaultApi.GetApplications(cmd.Context(), environment, nil)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			if httpResponse.Body != nil {
				defer httpResponse.Body.Close()
				respBody, parseError := io.ReadAll(httpResponse.Body)
				if parseError == nil {
					return fmt.Errorf("failed to list applications in the environment \"%s\": %s", environment, string(respBody))
				}
			}
		}
		return err
	}

	var list []cmfsdk.Application
	applications := append(list, applicationsPage.Items...)

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			jobStatus := app.Status["jobStatus"].(map[string]interface{})
			envInApp, ok := app.Metadata["environment"].(string)
			if !ok {
				envInApp = environment
			}
			list.Add(&flinkApplicationOut{
				Name:        app.Metadata["name"].(string),
				Environment: envInApp,
				JobId:       jobStatus["jobId"].(string),
				JobState:    jobStatus["state"].(string),
			})
		}
		return list.Print()
	}
	// if the output format is not human, we serialize the output as it is (JSON or YAML)
	return output.SerializedOutput(cmd, applications)
}
