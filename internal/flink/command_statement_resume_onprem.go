package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newStatementResumeCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "resume <statement-name>",
		Short:       "Resume a Flink SQL statement.",
		Long:        "Resume a Flink SQL statement in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementResumeOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementResumeOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	statement, err := client.GetStatement(c.createContext(), environment, name)
	if err != nil {
		return err
	}

	// Construct the statement to be stopped
	statement = cmfsdk.Statement{
		ApiVersion: statement.GetApiVersion(),
		Kind:       statement.GetKind(),
		Metadata: cmfsdk.StatementMetadata{
			Name: statement.GetMetadata().Name,
		},
		Spec: statement.GetSpec(),
	}

	statement.Spec.SetStopped(false)

	if err := client.UpdateStatement(c.createContext(), environment, name, statement); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "Requested to resume %s \"%s\".\n", resource.FlinkStatement, name)
	return nil
}
