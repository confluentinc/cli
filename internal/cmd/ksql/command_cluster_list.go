package ksql

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *ksqlCommand) newListCommand(resource string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List ksqlDB %ss.", resource),
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) list(cmd *cobra.Command, _ []string) error {
	clusters, err := c.V2Client.ListKsqlClusters(c.EnvironmentId())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters.Data {
		list.Add(c.formatClusterForDisplayAndList(&cluster))
	}
	return list.Print()
}
