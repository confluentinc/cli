package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithDefaultValue(cmd, output.JSON.String())

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationDescribe(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	// Disallow human output for this command
	if output.GetFormat(cmd) == output.Human {
		return errors.NewErrorWithSuggestions("human output is not supported for this command", "Try using --output flag with json or yaml.\n")
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applicationName := args[0]
	application, err := client.DescribeApplication(cmd.Context(), environment, applicationName)
	if err != nil {
		return err
	}

	return output.SerializedOutput(cmd, application)
}
