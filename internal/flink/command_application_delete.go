package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newApplicationDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink applications.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.applicationDelete,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	addCmfFlagSet(cmd)
	pcmd.AddForceFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationDelete(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, err := client.DescribeApplication(cmd.Context(), environment, name)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkApplication); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		return client.DeleteApplication(cmd.Context(), environment, name)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkApplication)
	return err
}
