package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applications, err := client.ListApplications(c.createContext(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			status := app.GetStatus()
			rawJobStatus, ok := status["jobStatus"]
			if !ok {
				return fmt.Errorf("job status not found in flink job status")
			}
			jobStatus, ok := rawJobStatus.(map[string]interface{})
			if !ok {
				return fmt.Errorf("jobStatus has unexpected type")
			}
			envInApp, ok := app.Spec["environment"].(string)
			if !ok {
				envInApp = environment
			}
			list.Add(&flinkApplicationSummaryOut{
				Name:        app.Metadata["name"].(string),
				Environment: envInApp,
				JobName:     jobStatus["jobName"].(string),
				JobStatus:   jobStatus["state"].(string),
			})
		}
		return list.Print()
	}

	localApps := make([]LocalFlinkApplication, 0, len(applications))

	for _, sdkApp := range applications {
		localApps = append(localApps, LocalFlinkApplication{
			ApiVersion: sdkApp.ApiVersion,
			Kind:       sdkApp.Kind,
			Metadata:   sdkApp.Metadata,
			Spec:       sdkApp.Spec,
			Status:     sdkApp.Status,
		})
	}

	return output.SerializedOutput(cmd, localApps)
}