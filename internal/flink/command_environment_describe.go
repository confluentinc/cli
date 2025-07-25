package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink environment.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.environmentDescribe,
	}

	addCmfFlagSet(cmd)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environmentName := args[0]
	sdkOutputEnvironment, err := client.DescribeEnvironment(c.createContext(), environmentName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printEnvironmentOutTable(cmd, sdkOutputEnvironment)
	}

	// Start with the top-level fields
localEnv := LocalEnvironment{
	Secrets:                  sdkOutputEnvironment.Secrets,
	Name:                     sdkOutputEnvironment.Name,
	CreatedTime:              sdkOutputEnvironment.CreatedTime,
	UpdatedTime:              sdkOutputEnvironment.UpdatedTime,
	FlinkApplicationDefaults: sdkOutputEnvironment.FlinkApplicationDefaults,
	KubernetesNamespace:      sdkOutputEnvironment.KubernetesNamespace,
	ComputePoolDefaults:      sdkOutputEnvironment.ComputePoolDefaults,
}

// Perform a deep copy for the nested StatementDefaults struct, handling nil pointers.
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

	return output.SerializedOutput(cmd, localEnv)
}
