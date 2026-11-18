package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newStatementEventListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <statement-name>",
		Short: "List events for a Flink SQL statement.",
		Long:  "List the events of a Flink SQL statement in Confluent Platform, most recent occurrence first. Requires CMF 2.5.0 or later.\n\nProblems with a shared compute pool, such as an image that can't be pulled, are recorded on the compute pool's events instead. List them with the `compute-pool event list` command.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.statementEventListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addPageSizeFlag(cmd)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementEventListOnPrem(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	pageSize, err := getPageSize(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	events, err := client.ListStatementEvents(c.createContext(), environment, args[0], pageSize)
	if err != nil {
		return err
	}

	return printResourceEvents(cmd, events)
}
