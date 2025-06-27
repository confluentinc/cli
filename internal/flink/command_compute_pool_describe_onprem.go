package flink

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newComputePoolDescribeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "describe <name>",
		Short:       "Describe a Flink compute pool in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.computePoolDescribeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) computePoolDescribeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	computePool, err := client.DescribeComputePool(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		table := output.NewTable(cmd)
		// nil pointer handling for creation timestamp
		var creationTime string
		if computePool.GetMetadata().CreationTimestamp != nil {
			creationTime = *computePool.GetMetadata().CreationTimestamp
		} else {
			creationTime = ""
		}

		table.Add(&computePoolOutOnPrem{
			CreationTime: creationTime,
			Name:         computePool.GetMetadata().Name,
			Type:         computePool.GetSpec().Type,
			Phase:        computePool.GetStatus().Phase,
		})
		return table.Print()
	}

	if output.GetFormat(cmd) == output.YAML {
		// Convert the computePool to our local struct for correct YAML field names
		jsonBytes, err := json.Marshal(computePool)
		if err != nil {
			return err
		}
		var outputLocalPool localComputePoolOnPrem
		if err = json.Unmarshal(jsonBytes, &outputLocalPool); err != nil {
			return err
		}
		// Output the local struct for correct YAML field names
		out, err := yaml.Marshal(outputLocalPool)
		if err != nil {
			return err
		}
		output.Print(false, string(out))
		return nil
	}

	return output.SerializedOutput(cmd, computePool)
}
