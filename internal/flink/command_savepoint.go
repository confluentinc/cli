package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type savepointOut struct {
	Name         string `human:"Name" serialized:"name"`
	Application  string `human:"Application,omitempty" serialized:"application,omitempty"`
	Statement    string `human:"Statement,omitempty" serialized:"statement,omitempty"`
	Path         string `human:"Path,omitempty" serialized:"path,omitempty"`
	Format       string `human:"Format,omitempty" serialized:"format,omitempty"`
	BackoffLimit int32  `human:"Backoff Limit,omitempty" serialized:"backoff_limit,omitempty"`
	Uid          string `human:"UID,omitempty" serialized:"uid,omitempty"`
	State        string `human:"State,omitempty" serialized:"state,omitempty"`
}

func (c *command) newSavepointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "savepoint",
		Short:       "Manage Flink savepoints.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newSavepointCreateCommand())
	cmd.AddCommand(c.newSavepointDescribeCommand())
	cmd.AddCommand(c.newSavepointDetachCommand())
	cmd.AddCommand(c.newSavepointDeleteCommand())
	cmd.AddCommand(c.newSavepointListCommand())

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
			State:            sdkSavepoint.Status.State,
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
