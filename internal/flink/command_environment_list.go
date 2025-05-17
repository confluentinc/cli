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

	environments, err := client.ListEnvironments(c.createContext())
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		list.Filter([]string{"Name", "CreatedTime", "UpdatedTime", "KubernetesNamespace"})
		for _, env := range environments {
			list.Add(&flinkEnvironmentOutput{
				Name:                env.Name,
				KubernetesNamespace: env.KubernetesNamespace,
				CreatedTime:         env.CreatedTime.String(),
				UpdatedTime:         env.UpdatedTime.String(),
			})
		}
		return list.Print()
	}
	return output.SerializedOutput(cmd, environments)
}
