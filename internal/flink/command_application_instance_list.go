package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationInstanceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink application instances.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationInstanceList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "Name of the Flink application.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("application"))

	return cmd
}

func (c *command) applicationInstanceList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	instances, err := client.ListApplicationInstances(c.createContext(), environment, application)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, instance := range instances {
			metadata := instance.GetMetadata()
			var creationTime string
			if metadata.CreationTimestamp != nil {
				creationTime = *metadata.CreationTimestamp
			}

			var jobId, jobState string
			if instance.Status != nil && instance.Status.JobStatus != nil {
				if instance.Status.JobStatus.JobId != nil {
					jobId = *instance.Status.JobStatus.JobId
				}
				if instance.Status.JobStatus.State != nil {
					jobState = *instance.Status.JobStatus.State
				}
			}

			list.Add(&flinkApplicationInstanceSummaryOut{
				Name:         metadata.GetName(),
				CreationTime: creationTime,
				JobId:        jobId,
				JobState:     jobState,
			})
		}
		return list.Print()
	}

	localInstances := make([]LocalFlinkApplicationInstance, 0, len(instances))
	for _, sdkInstance := range instances {
		localInstances = append(localInstances, convertSdkApplicationInstanceToLocalApplicationInstance(sdkInstance))
	}

	return output.SerializedOutput(cmd, localInstances)
}
