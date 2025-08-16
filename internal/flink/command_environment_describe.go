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

	localEnv := convertSdkEnvironmentToLocalEnvironment(sdkOutputEnvironment)
	return output.SerializedOutput(cmd, localEnv)
}
