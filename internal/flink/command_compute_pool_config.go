package flink

import (
	"github.com/spf13/cobra"
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
