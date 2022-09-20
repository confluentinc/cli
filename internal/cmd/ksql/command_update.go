package ksql

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
)

func (c *ksqlCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing ksqlDB cluster.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
	}

	cmd.Flags().Int("csu", 0, "Number of Confluent Streaming Units (non-negative) requested for ksqlDB cluster.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("csu")

	return cmd
}

func (c *ksqlCommand) update(cmd *cobra.Command, args []string) error {
	id := args[0]

	req := &schedv1.KSQLCluster{
		AccountId: c.EnvironmentId(),
		Id:        id,
	}

	// Check KSQL exists
	cluster, err := c.Client.KSQL.Describe(context.Background(), req)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, id)
	}

	if cluster.Status == schedv1.ClusterStatus_UP {
		csu, err := cmd.Flags().GetInt("csu")
		if err != nil {
			return err
		}
		err = c.validateCsu(csu)
		if err != nil {
			return err
		}
		cluster.TotalNumCsu = uint32(csu)
		_, err = c.Client.KSQL.Update(context.Background(), cluster)
		if err != nil {
			return err
		}
	}

	return err
}

func (c *ksqlCommand) validateCsu(csu int) error {
	if csu != 1 && csu != 2 && csu != 4 && csu != 8 && csu != 12 {
		return errors.Errorf("failed to update ksql cluster: %v", errors.CSUInvalidErrorMsg)
	}
	return nil
}
