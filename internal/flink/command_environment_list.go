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

	printableEnvs := make([]LocalEnvironment, 0, len(sdkEnvironments))

	for _, sdkEnv := range sdkEnvironments {

		localEnv := LocalEnvironment{
			Secrets:                  sdkEnv.Secrets,
			Name:                     sdkEnv.Name,
			CreatedTime:              sdkEnv.CreatedTime,
			UpdatedTime:              sdkEnv.UpdatedTime,
			FlinkApplicationDefaults: sdkEnv.FlinkApplicationDefaults,
			KubernetesNamespace:      sdkEnv.KubernetesNamespace,
			ComputePoolDefaults:      sdkEnv.ComputePoolDefaults,
		}

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

		printableEnvs = append(printableEnvs, localEnv)
	}

	return output.SerializedOutput(cmd, printableEnvs)
}
