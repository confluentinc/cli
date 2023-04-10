package flink

import (
	"github.com/spf13/cobra"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	pproperties "github.com/confluentinc/cli/internal/pkg/properties"
)

func (c *command) newStatementCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <sql>",
		Short: "Create a Flink SQL statement.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.statementCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a Flink SQL statement.",
				Code: `confluent flink statement create "SELECT * FROM table;"`,
			},
		),
	}

	cmd.Flags().String("name", "", "The name of the Flink SQL statement.")
	cmd.Flags().String("compute-pool", "", "Flink compute pool ID.")
	cmd.Flags().StringSlice("config", []string{}, `A comma-separated list of configuration "key=value" pairs.`)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) statementCreate(cmd *cobra.Command, args []string) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	computePool, err := cmd.Flags().GetString("compute-pool")
	if err != nil {
		return err
	}

	configs, err := cmd.Flags().GetStringSlice("config")
	if err != nil {
		return err
	}

	properties, err := pproperties.ConfigFlagToMap(configs)
	if err != nil {
		return err
	}

	statement := flinkgatewayv1alpha1.SqlV1alpha1Statement{Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
		StatementName: flinkgatewayv1alpha1.PtrString(name),
		Statement:     flinkgatewayv1alpha1.PtrString(args[0]),
		Properties:    &properties,
		ComputePoolId: flinkgatewayv1alpha1.PtrString(computePool),
	}}

	statement, err = c.V2Client.CreateStatement(c.EnvironmentId(), statement)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&statementOut{
		Name:         statement.Spec.GetStatementName(),
		Statement:    statement.Spec.GetStatement(),
		ComputePool:  statement.Spec.GetComputePoolId(),
		Status:       statement.Status.GetPhase(),
		StatusDetail: statement.Status.GetDetail(),
	})
	return table.Print()
}
