package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newStatementRescaleCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "rescale <statement-name>",
		Short:       "Rescale a Flink SQL statement in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementRescaleOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().Int32("parallelism", 4, "New parallelism of the Flink SQL statement.")
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("parallelism"))

	return cmd
}

func (c *command) statementRescaleOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	parallelism, err := cmd.Flags().GetInt32("parallelism")
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	// Construct the statement to be stopped first
	statement = cmfsdk.Statement{
		ApiVersion: statement.GetApiVersion(),
		Kind:       statement.GetKind(),
		Metadata: cmfsdk.StatementMetadata{
			Name: statement.GetMetadata().Name,
		},
		Spec: statement.GetSpec(),
	}

	statement.Spec.SetStopped(true)

	if err = client.UpdateStatement(c.createContext(), environment, name, statement); err != nil {
		return err
	}

	// Read the statement again to get the latest state
	statement, err = client.GetStatement(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	// Construct the statement to resume later with different parallelism
	statement = cmfsdk.Statement{
		ApiVersion: statement.GetApiVersion(),
		Kind:       statement.GetKind(),
		Metadata: cmfsdk.StatementMetadata{
			Name: statement.GetMetadata().Name,
		},
		Spec: statement.GetSpec(),
	}

	statement.Spec.SetStopped(false)
	statement.Spec.SetParallelism(parallelism)

	if err = client.UpdateStatement(c.createContext(), environment, name, statement); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "Requested to rescale %s \"%s\" with new parallelism = %d.\n", resource.FlinkStatement, name, parallelism)
	return nil
}
