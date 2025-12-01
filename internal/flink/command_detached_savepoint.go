package flink

import (
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

type detachedSavepointOut struct {
	Name              string `human:"Name" serialized:"name"`
	Path              string `human:"Path,omitempty" serialized:"path,omitempty"`
	Format            string `human:"Format,omitempty" serialized:"format,omitempty"`
	Limit             int32  `human:"Backoff Limit,omitempty" serialized:"backoff_limit,omitempty"`
	CreationTimestamp string `human:"Creation Timestamp,omitempty" serialized:"creation_timestamp,omitempty"`
	Uid               string `human:"Uid,omitempty" serialized:"uid,omitempty"`
}

func (c *command) newDetachedSavepointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "detached-savepoint",
		Short:       "Manage Flink detached savepoints.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newDetachedSavepointCreateCommand())
	cmd.AddCommand(c.newDetachedSavepointDescribeCommand())
	cmd.AddCommand(c.newDetachedSavepointListCommand())
	cmd.AddCommand(c.newDetachedSavepointDeleteCommand())

	return cmd
}

func convertSdkDetachedSavepointToLocalSavepoint(sdkSavepoint cmfsdk.Savepoint) LocalSavepoint {
	localSavepoint := LocalSavepoint{
		ApiVersion: sdkSavepoint.ApiVersion,
		Kind:       sdkSavepoint.Kind,
		Metadata: LocalSavepointMetadata{
			Name:              *sdkSavepoint.Metadata.Name,
			CreationTimestamp: sdkSavepoint.Metadata.CreationTimestamp,
			Uid:               sdkSavepoint.Metadata.Uid,
			Labels:            sdkSavepoint.Metadata.Labels,
			Annotations:       sdkSavepoint.Metadata.Annotations,
		},
		Spec: LocalSavepointSpec{
			Path:         sdkSavepoint.Spec.Path,
			BackoffLimit: sdkSavepoint.Spec.BackoffLimit,
			FormatType:   sdkSavepoint.Spec.FormatType,
		},
	}

	if sdkSavepoint.Status != nil {
		localSavepoint.Status = &LocalSavepointStatus{
			State:            sdkSavepoint.Status.Path,
			Path:             sdkSavepoint.Status.Path,
			TriggerTimestamp: sdkSavepoint.Status.TriggerTimestamp,
			ResultTimestamp:  sdkSavepoint.Status.ResultTimestamp,
			Failures:         sdkSavepoint.Status.Failures,
			Error:            sdkSavepoint.Status.Error,
			PendingDeletion:  sdkSavepoint.Status.PendingDeletion,
		}
	}

	return localSavepoint
}
