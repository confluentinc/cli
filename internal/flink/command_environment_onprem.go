package flink

import (
	"github.com/spf13/cobra"
)

type flinkEnvironmentSummary struct {
	Name        string `human:"Name" serialized:"name"`
	CreatedTime string `human:"Created Time" serialized:"created_time"`
	UpdatedTime string `human:"Updated Time" serialized:"updated_time"`
}

type flinkEnvironmentOutput struct {
	Name                     string `human:"Name" serialized:"name"`
	KubernetesNamespace      string `human:"Kubernetes Namespace" serialized:"kubernetes_namespace"`
	CreatedTime              string `human:"Created Time" serialized:"created_time"`
	UpdatedTime              string `human:"Updated Time" serialized:"updated_time"`
	FlinkApplicationDefaults string `human:"Flink Application Defaults" serialized:"flinkApplicationDefaults"`
}

func (c *command) newEnvironmentCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "environment",
		Short:   "Manage Flink environments",
		Aliases: []string{"env"},
	}

	cmd.AddCommand(c.newEnvironmentCreateCommand())
	cmd.AddCommand(c.newEnvironmentDeleteCommand())
	cmd.AddCommand(c.newEnvironmentDescribeCommand())
	cmd.AddCommand(c.newEnvironmentListCommand())
	cmd.AddCommand(c.newEnvironmentUpdateCommand())
	cmd.AddCommand(c.newEnvironmentUseCommand())
	return cmd
}
