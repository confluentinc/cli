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

	cmd.Flags().String("environment", "", "Name of the environment to rescale the Flink SQL statement.")
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

	// Construct the statement to be stopped first
	statementToStop := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Stopped: cmfsdk.PtrBool(true),
		},
	}

	if err = client.UpdateStatement(c.createContext(), environment, name, statementToStop); err != nil {
		return err
	}

	// Construct the statement to resume later with different parallelism
	statementToResume := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Stopped:     cmfsdk.PtrBool(false),
			Parallelism: cmfsdk.PtrInt32(parallelism),
		},
	}

	if err = client.UpdateStatement(c.createContext(), environment, name, statementToResume); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "Requested to rescale %s \"%s\" with new parallelism = %d.\n", resource.FlinkStatement, name, parallelism)
	return nil
}
