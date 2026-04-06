package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type flinkApplicationInstanceSummaryOut struct {
	Name          string `human:"Name" serialized:"name"`
	CreationTime  string `human:"Creation Time" serialized:"creation_time"`
	JobId         string `human:"Job ID" serialized:"job_id"`
	JobState      string `human:"Job State" serialized:"job_state"`
}

func (c *command) newApplicationInstanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "instance",
		Short:       "Manage Flink application instances.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationInstanceListCommand())

	return cmd
}

func convertSdkApplicationInstanceToLocalApplicationInstance(sdkInstance cmfsdk.FlinkApplicationInstance) LocalFlinkApplicationInstance {
	localInstance := LocalFlinkApplicationInstance{
		ApiVersion: sdkInstance.ApiVersion,
		Kind:       sdkInstance.Kind,
	}
	if sdkInstance.Metadata != nil {
		localInstance.Metadata = &LocalApplicationInstanceMetadata{
			Name:              sdkInstance.Metadata.Name,
			Uid:               sdkInstance.Metadata.Uid,
			CreationTimestamp: sdkInstance.Metadata.CreationTimestamp,
			UpdateTimestamp:   sdkInstance.Metadata.UpdateTimestamp,
			Labels:            sdkInstance.Metadata.Labels,
			Annotations:       sdkInstance.Metadata.Annotations,
		}
	}
	if sdkInstance.Status != nil {
		localInstance.Status = &LocalApplicationInstanceStatus{
			Spec: sdkInstance.Status.Spec,
		}
		if sdkInstance.Status.JobStatus != nil {
			localInstance.Status.JobStatus = &LocalApplicationInstanceJobStatus{
				JobId: sdkInstance.Status.JobStatus.JobId,
				State: sdkInstance.Status.JobStatus.State,
			}
		}
	}
	return localInstance
}
