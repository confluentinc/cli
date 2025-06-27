package flink

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlagWithHumanRestricted(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationDescribe(cmd *cobra.Command, args []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applicationName := args[0]
	application, err := client.DescribeApplication(c.createContext(), environment, applicationName)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.YAML {
		// Convert the application to our local struct for correct YAML field names
		jsonBytes, err := json.Marshal(application)
		if err != nil {
			return err
		}
		var outputLocalApp localFlinkApplication
		if err = json.Unmarshal(jsonBytes, &outputLocalApp); err != nil {
			return err
		}
		// Output the local struct for correct YAML field names
		out, err := yaml.Marshal(outputLocalApp)
		if err != nil {
			return err
		}
		output.Print(false, string(out))
		return nil
	}

	return output.SerializedOutput(cmd, application)
}
