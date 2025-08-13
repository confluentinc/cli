package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <resourceFilePath>",
		Short: "Update a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationUpdate,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationUpdate(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	sdkApplication, err := readApplicationResourceFile(args[0])
	if err != nil {
		return err
	}

	sdkOutputApplication, err := client.UpdateApplication(c.createContext(), environment, sdkApplication)
	if err != nil {
		return err
	}

	localOutputApp := convertSdkApplicationToLocalApplication(sdkOutputApplication)

	return output.SerializedOutput(cmd, localOutputApp)
}
