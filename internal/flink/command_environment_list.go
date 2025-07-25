package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newEnvironmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink environments.",
		Args:  cobra.NoArgs,
		RunE:  c.environmentList,
	}

	addCmfFlagSet(cmd)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentList(cmd *cobra.Command, _ []string) error {
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkEnvironments, err := client.ListEnvironments(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		list.Filter([]string{"Name", "CreatedTime", "UpdatedTime", "KubernetesNamespace"})
		for _, env := range sdkEnvironments {
			list.Add(&flinkEnvironmentOutput{
				Name:                env.Name,
				KubernetesNamespace: env.KubernetesNamespace,
				CreatedTime:         env.CreatedTime.String(),
				UpdatedTime:         env.UpdatedTime.String(),
			})
		}
		return list.Print()
	}

	// Create the slice to hold the clean objects
printableEnvs := make([]LocalEnvironment, 0, len(sdkEnvironments))

// Loop through the original SDK objects
for _, sdkEnv := range sdkEnvironments {

	// --- Start Deep Copy for each item ---

	// Start with the top-level fields
	localEnv := LocalEnvironment{
		Secrets:                  sdkEnv.Secrets,
		Name:                     sdkEnv.Name,
		CreatedTime:              sdkEnv.CreatedTime,
		UpdatedTime:              sdkEnv.UpdatedTime,
		FlinkApplicationDefaults: sdkEnv.FlinkApplicationDefaults,
		KubernetesNamespace:      sdkEnv.KubernetesNamespace,
		ComputePoolDefaults:      sdkEnv.ComputePoolDefaults,
	}

	// Perform a deep copy for the nested StatementDefaults struct, handling nil pointers.
	if sdkEnv.StatementDefaults != nil {
		localDefaults1 := &LocalAllStatementDefaults1{}

		if sdkEnv.StatementDefaults.Detached != nil {
			localDefaults1.Detached = &LocalStatementDefaults{
				FlinkConfiguration: sdkEnv.StatementDefaults.Detached.FlinkConfiguration,
			}
		}

		if sdkEnv.StatementDefaults.Interactive != nil {
			localDefaults1.Interactive = &LocalStatementDefaults{
				FlinkConfiguration: sdkEnv.StatementDefaults.Interactive.FlinkConfiguration,
			}
		}

		localEnv.StatementDefaults = localDefaults1
	}

	// Append the fully "clean" object to the final slice
	printableEnvs = append(printableEnvs, localEnv)
}

// Now, printableEnvs is ready to be serialized
return output.SerializedOutput(cmd, printableEnvs)

	return output.SerializedOutput(cmd, printableEnvs)
}
