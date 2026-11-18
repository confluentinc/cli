package flink

import (
	"time"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type flinkEnvironmentOutput struct {
	Name                         string `human:"Name" serialized:"name"`
	KubernetesNamespace          string `human:"Kubernetes Namespace" serialized:"kubernetes_namespace"`
	CreatedTime                  string `human:"Created Time" serialized:"created_time"`
	UpdatedTime                  string `human:"Updated Time" serialized:"updated_time"`
	DeletionRequestedTime        string `human:"Deletion Requested Time,omitempty" serialized:"deletion_requested_time,omitempty"`
	FlinkApplicationDefaults     string `human:"Flink Application Defaults,omitempty" serialized:"flink_application_defaults,omitempty"`
	ComputePoolDefaults          string `human:"Flink Compute Pool Defaults,omitempty" serialized:"flink_compute_pool_defaults,omitempty"`
	DetachedStatementDefaults    string `human:"Detached Statement Defaults,omitempty" serialized:"detached_statement_defaults,omitempty"`
	InteractiveStatementDefaults string `human:"Interactive Statement Defaults,omitempty" serialized:"interactive_statement_defaults,omitempty"`
}

func (c *command) newEnvironmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "environment",
		Short:       "Manage Flink environments.",
		Aliases:     []string{"env"},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newEnvironmentCreateCommand())
	cmd.AddCommand(c.newEnvironmentDeleteCommand())
	cmd.AddCommand(c.newEnvironmentDescribeCommand())
	cmd.AddCommand(c.newEnvironmentListCommand())
	cmd.AddCommand(c.newEnvironmentUpdateCommand())

	return cmd
}

func convertSdkEnvironmentToLocalEnvironment(sdkOutputEnvironment cmfsdk.Environment) LocalEnvironment {
	localEnv := LocalEnvironment{
		Secrets:                  sdkOutputEnvironment.Secrets,
		Name:                     sdkOutputEnvironment.Name,
		CreatedTime:              sdkOutputEnvironment.CreatedTime,
		UpdatedTime:              sdkOutputEnvironment.UpdatedTime,
		FlinkApplicationDefaults: sdkOutputEnvironment.FlinkApplicationDefaults,
		KubernetesNamespace:      sdkOutputEnvironment.KubernetesNamespace,
		ComputePoolDefaults:      sdkOutputEnvironment.ComputePoolDefaults,
	}

	if sdkOutputEnvironment.StatementDefaults != nil {
		localDefaults1 := &LocalAllStatementDefaults1{}

		if sdkOutputEnvironment.StatementDefaults.Detached != nil {
			localDefaults1.Detached = &LocalStatementDefaults{
				FlinkConfiguration: sdkOutputEnvironment.StatementDefaults.Detached.FlinkConfiguration,
			}
		}

		if sdkOutputEnvironment.StatementDefaults.Interactive != nil {
			localDefaults1.Interactive = &LocalStatementDefaults{
				FlinkConfiguration: sdkOutputEnvironment.StatementDefaults.Interactive.FlinkConfiguration,
			}
		}

		localEnv.StatementDefaults = localDefaults1
	}

	if timestamp := sdkOutputEnvironment.Metadata.GetDeletionRequestedTimestamp(); timestamp != "" {
		localEnv.Metadata = &LocalEnvironmentMetadata{DeletionRequestedTimestamp: timestamp}
	}

	return localEnv
}

// formatDeletionRequestedTime formats when the environment's deletion was requested the way its created and updated
// times are shown. CMF's spec types the timestamp as a plain string, so a value that isn't RFC 3339 is shown as it is.
func formatDeletionRequestedTime(metadata *cmfsdk.EnvironmentMetadata) string {
	timestamp := metadata.GetDeletionRequestedTimestamp()
	if t, err := time.Parse(time.RFC3339, timestamp); err == nil {
		return t.String()
	}
	return timestamp
}
