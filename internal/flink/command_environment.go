package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type flinkEnvironmentOutput struct {
	Name                     string `human:"Name" serialized:"name"`
	KubernetesNamespace      string `human:"Kubernetes Namespace" serialized:"kubernetes_namespace"`
	CreatedTime              string `human:"Created Time" serialized:"created_time"`
	UpdatedTime              string `human:"Updated Time" serialized:"updated_time"`
	FlinkApplicationDefaults string `human:"Flink Application Defaults" serialized:"flink_application_defaults"`
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
