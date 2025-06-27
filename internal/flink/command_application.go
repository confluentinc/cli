package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type flinkApplicationSummaryOut struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	JobName     string `human:"Job Name" serialized:"job_name"`
	JobStatus   string `human:"Job Status" serialized:"job_status"`
}

// localFlinkApplication is a local struct with YAML tags that matches the SDK FlinkApplication structure
type localFlinkApplication struct {
	ApiVersion string                  `yaml:"apiVersion" json:"apiVersion"`
	Kind       string                  `yaml:"kind" json:"kind"`
	Metadata   map[string]interface{}  `yaml:"metadata" json:"metadata"`
	Spec       map[string]interface{}  `yaml:"spec" json:"spec"`
	Status     *map[string]interface{} `yaml:"status,omitempty" json:"status,omitempty"`
}

func (c *command) newApplicationCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "application",
		Short:       "Manage Flink applications.",
		Aliases:     []string{"app"},
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
	}

	cmd.AddCommand(c.newApplicationCreateCommand())
	cmd.AddCommand(c.newApplicationDeleteCommand())
	cmd.AddCommand(c.newApplicationDescribeCommand())
	cmd.AddCommand(c.newApplicationListCommand())
	cmd.AddCommand(c.newApplicationUpdateCommand())
	cmd.AddCommand(c.newApplicationWebUiForwardCommand())

	return cmd
}
