package flink

import (
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newStatementStopCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "stop <statement-name>",
		Short:       "Stop a Flink SQL statement in Confluent Platform.",
		Args:        cobra.ExactArgs(1),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogout},
		RunE:        c.statementStopOnPrem,
	}

	cmd.Flags().String("environment", "", "Name of the environment to stop the Flink SQL statement.")
	addCmfFlagSet(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) statementStopOnPrem(cmd *cobra.Command, args []string) error {
	name := args[0]

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	// Construct the statement to be stopped
	// TODO: Check with Fabian if this is enough or not
	statement := cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: name,
		},
		Spec: cmfsdk.StatementSpec{
			Stopped: true,
		},
	}

	if err := client.UpdateStatement(c.createContext(), environment, name, statement); err != nil {
		return err
	}
	output.Printf(c.Config.EnableColor, "Requested to stop %s \"%s\".\n", resource.FlinkStatement, name)
	return nil
}
