package ksql

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *ksqlCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: fmt.Sprintf("List ksqlDB clusters."),
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ksqlCommand) list(cmd *cobra.Command, _ []string) error {
	req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
	clusters, err := c.Client.KSQL.List(context.Background(), req)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, cluster := range clusters {
		list.Add(c.updateKsqlClusterForDescribeAndList(cluster))
	}
	return list.Print()
}
