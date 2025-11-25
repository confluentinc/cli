package flink

import (
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

type savepointOut struct {
	Name        string `human:"Name" serialized:"name"`
	Application string `human:"Application,omitempty" serialized:"application,omitempty"`
	Statement   string `human:"Statement,omitempty" serialized:"statement,omitempty"`
	Path        string `human:"Path,omitempty" serialized:"path,omitempty"`
	Format      string `human:"Format,omitempty" serialized:"format,omitempty"`
	Limit       int32  `human:"Backoff Limit,omitempty" serialized:"backoff_limit,omitempty"`
}

func (c *command) newSavepointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "savepoint",
		Short: "Manage Flink savepoint.",
	}

	cmd.AddCommand(c.newSavepointCreateCommand())
	cmd.AddCommand(c.newSavepointDescribeCommand())
	cmd.AddCommand(c.newSavepointListCommand())
	cmd.AddCommand(c.newSavepointDetachCommand())
	cmd.AddCommand(c.newSavepointDeleteCommand())

	return cmd
}

func convertSdkSavepointToLocalSavepoint(sdkSavepoint cmfsdk.Savepoint) LocalSavepoint {
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
