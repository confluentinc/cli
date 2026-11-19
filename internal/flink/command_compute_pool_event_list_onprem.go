package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *command) newComputePoolEventListCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list <compute-pool-name>",
		Short: "List events for a Flink compute pool.",
		Long:  "List the events of a Flink compute pool in Confluent Platform, most recent occurrence first. Requires CMF 2.5.0 or later.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.computePoolEventListOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addPageSizeFlag(cmd)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolEventListOnPrem(cmd *cobra.Command, args []string) error {
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

	events, err := client.ListComputePoolEvents(c.createContext(), environment, args[0], pageSize)
	if err != nil {
		return err
	}

	return printResourceEvents(cmd, events)
}
