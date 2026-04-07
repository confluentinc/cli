package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationInstanceDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink application instance.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationInstanceDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("application", "", "Name of the Flink application.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("application"))

	return cmd
}

func (c *command) applicationInstanceDescribe(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	application, err := cmd.Flags().GetString("application")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	instanceName := args[0]
	instance, err := client.DescribeApplicationInstance(c.createContext(), environment, application, instanceName)
	if err != nil {
		return err
	}

	localInstance := convertSdkApplicationInstanceToLocalApplicationInstance(instance)

	return output.SerializedOutput(cmd, localInstance)
}
