package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

type computePoolOut struct {
	IsCurrent   bool   `human:"Current" serialized:"is_current"`
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	CurrentCfu  int32  `human:"Current CFU" serialized:"currrent_cfu"`
	MaxCfu      int32  `human:"Max CFU" serialized:"max_cfu"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	Status      string `human:"Status" serialized:"status"`
}

type computePoolOutOnPrem struct {
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	ID          string `human:"ID" serialized:"id"`
	Phase       string `human:"Phase" serialized:"phase"`
}

func (c *command) newComputePoolCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute-pool",
		Short: "Manage Flink compute pools.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newComputePoolCreateCommand())
		cmd.AddCommand(c.newComputePoolDeleteCommand())
		cmd.AddCommand(c.newComputePoolDescribeCommand())
		cmd.AddCommand(c.newComputePoolListCommand())
		cmd.AddCommand(c.newComputePoolUnsetCommand())
		cmd.AddCommand(c.newComputePoolUpdateCommand())
		cmd.AddCommand(c.newComputePoolUseCommand())
	} else {
		cmd.AddCommand(c.newComputePoolCreateCommandOnPrem())
		cmd.AddCommand(c.newComputePoolDescribeCommandOnPrem())
		cmd.AddCommand(c.newComputePoolListCommandOnPrem())
		cmd.AddCommand(c.newComputePoolDeleteCommandOnPrem())
	}

	return cmd
}

func (c *command) validComputePoolArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.autocompleteComputePools(cmd, args)
}
