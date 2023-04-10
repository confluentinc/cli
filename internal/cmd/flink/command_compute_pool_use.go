package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newComputePoolUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use <id>",
		Short: "Choose a Flink compute pool to be used in subsequent commands.",
		Long:  "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.use,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) use(cmd *cobra.Command, args []string) error {
	context := c.Config.Contexts[c.Config.CurrentContext]
	context.Environments[context.CurrentEnvironment].CurrentFlinkComputePool = args[0]
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf("Using compute pool \"%s\".\n", args[0])
	return nil
}
