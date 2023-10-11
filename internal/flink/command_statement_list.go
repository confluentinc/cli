package flink

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

var allowedStatuses = []string{
	"PENDING",
	"RUNNING",
	"COMPLETED",
	"DELETING",
	"FAILING",
	"FAILED",
	"STOPPED",
}

func (c *command) newStatementListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		RunE:  c.statementList,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.Flags().String("status", "", "Filter the results by statement status")
	pcmd.RegisterFlagCompletionFunc(cmd, "status", func(*cobra.Command, []string) []string {
		return allowedStatuses
	})

	return cmd
}

func (c *command) statementList(cmd *cobra.Command, args []string) error {
	client, err := c.GetFlinkGatewayClient(false)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return err
	}
	status = strings.ToUpper(status)

	if status != "" && !slices.Contains(allowedStatuses, status) {
		return errors.NewErrorWithSuggestions(
			"invalid value for flag --status",
			fmt.Sprintf("Please select a value from the following: [%s]", strings.Join(allowedStatuses, ", ")),
		)
	}

	statements, err := client.ListAllStatements(environmentId, c.Context.GetCurrentOrganization(), c.Context.GetCurrentFlinkComputePool())
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, statement := range statements {
		// filter out statements with a non-matching status
		if status != "" && statement.Status.GetPhase() != status {
			continue
		}

		list.Add(&statementOut{
			CreationDate: statement.Metadata.GetCreatedAt(),
			Name:         statement.GetName(),
			Statement:    statement.Spec.GetStatement(),
			ComputePool:  statement.Spec.GetComputePoolId(),
			Status:       statement.Status.GetPhase(),
			StatusDetail: statement.Status.GetDetail(),
		})
	}
	return list.Print()
}
