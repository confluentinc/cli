package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func (c *command) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
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

	applications, err := client.ListApplications(cmd.Context(), environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			appSummary := populateFlinkApplicationSummaryOut(app, environment)
			list.Add(appSummary)
		}
		return list.Print()
	}
	// if the output format is not human, we serialize the output as it is (JSON or YAML)
	return output.SerializedOutput(cmd, applications)
}

func populateFlinkApplicationSummaryOut(application cmfsdk.Application, envFromFlag string) *flinkApplicationSummaryOut {
	var appSummary *flinkApplicationSummaryOut

	var jobStatus map[string]any = getOrDefault(application.Status, "jobStatus", map[string]any{})
	jobNameString := getOrDefault(jobStatus, "jobName", "")
	jobStatusString := getOrDefault(jobStatus, "state", "")
	name := getOrDefault(application.Metadata, "name", "")
	environment := getOrDefault(application.Spec, "environment", envFromFlag)

	appSummary = &flinkApplicationSummaryOut{
		Name:        name,
		Environment: environment,
		JobName:     jobNameString,
		JobStatus:   jobStatusString,
	}

	return appSummary
}

func getOrDefault[T any](m map[string]any, key string, d T) T {
	value, ok := m[key]
	if !ok {
		return d
	}
	valueCast, ok := value.(T)
	if !ok {
		return d
	}
	return valueCast
}
