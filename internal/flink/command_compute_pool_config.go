package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type computePoolConfigOut struct {
	DefaultPoolEnabled bool  `human:"Default Pool Enabled" serialized:"default_pool_enabled"`
	DefaultPoolMaxCFU  int32 `human:"Default Pool Max CFU" serialized:"default_pool_max_cfu"`
}

func (c *command) newComputePoolConfigCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool-config",
		Short: "Manage Flink compute pools configs.",
	}

	cmd.AddCommand(c.newComputePoolConfigDescribeCommand())
	cmd.AddCommand(c.newComputePoolConfigUpdateCommand())

	return cmd
}

// validComputePoolArgs completes this command's optional compute pool positional. It used to
// borrow the identically named method from command_compute_pool.go; that file is now generated
// and its copy lives on the generated *computePoolCommand receiver, so this command needs its
// own. The query itself is the shared pcmd helper.
func (c *command) validComputePoolArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}
	return pcmd.AutocompleteComputePools(environmentId, c.V2Client)
}
