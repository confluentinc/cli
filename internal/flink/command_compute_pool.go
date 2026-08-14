package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

type computePoolOut struct {
	IsCurrent   bool   `human:"Current" serialized:"is_current"`
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name" serialized:"name"`
	Environment string `human:"Environment" serialized:"environment"`
	CurrentCfu  int32  `human:"Current CFU" serialized:"currrent_cfu"`
	MaxCfu      int32  `human:"Max CFU" serialized:"max_cfu"`
	DefaultPool bool   `human:"Default Pool" serialized:"default_pool"`
	Cloud       string `human:"Cloud" serialized:"cloud"`
	Region      string `human:"Region" serialized:"region"`
	Status      string `human:"Status" serialized:"status"`
}

func (c *command) newComputePoolCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "compute-pool",
		Short:       "Manage Flink compute pools in Confluent Cloud.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	cmd.AddCommand(c.newComputePoolCreateCommand())
	cmd.AddCommand(c.newComputePoolDeleteCommand())
	cmd.AddCommand(c.newComputePoolDescribeCommand())
	cmd.AddCommand(c.newComputePoolListCommand())
	cmd.AddCommand(c.newComputePoolUnsetCommand())
	cmd.AddCommand(c.newComputePoolUpdateCommand())
	cmd.AddCommand(c.newComputePoolUseCommand())

	return cmd
}

func (c *command) validComputePoolArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validComputePoolArgsMultiple(cmd, args)
}

func (c *command) validComputePoolArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteComputePools()
}

// Positional <id> completion, as opposed to the --compute-pool flag completion that
// pcmd.AddComputePoolFlag owns. Deliberately self-contained rather than delegating to
// pcmd.AutocompleteComputePools: this is the shape the generator emits, so keeping it here
// means generation replaces this file without changing behavior.
func (c *command) autocompleteComputePools() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	computePools, err := c.V2Client.ListFlinkComputePools(environmentId, "")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(computePools))
	for i, computePool := range computePools {
		spec := computePool.GetSpec()
		suggestions[i] = fmt.Sprintf("%s\t%s", computePool.GetId(), spec.GetDisplayName())
	}
	return suggestions
}
