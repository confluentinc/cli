package flink

import (
	"encoding/json"
	"fmt"

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

	// Get the name of the environment to be retrieved
	environmentName := args[0]
	environment, err := client.DescribeEnvironment(c.createContext(), environmentName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		var defaultsBytes []byte
		defaultsBytes, err = json.Marshal(environment.FlinkApplicationDefaults)
		if err != nil {
			return fmt.Errorf(`failed to marshal defaults for environment "%s": %s`, environmentName, err)
		}

		table.Add(&flinkEnvironmentOutput{
			Name:                     environment.Name,
			KubernetesNamespace:      environment.KubernetesNamespace,
			FlinkApplicationDefaults: string(defaultsBytes),
			CreatedTime:              environment.CreatedTime.String(),
			UpdatedTime:              environment.UpdatedTime.String(),
		})
		return table.Print()
	}
	return output.SerializedOutput(cmd, environment)
}
