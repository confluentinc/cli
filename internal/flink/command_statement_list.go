package flink

import (
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

var allowedStatuses = []string{
	"pending",
	"running",
	"completed",
	"deleting",
	"failing",
	"failed",
	"stopped",
}

func (c *command) newStatementListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		Example: examples.BuildExampleString(examples.Example{
			Text: "List running statements.",
			Code: "confluent flink statement list --status running",
		}),
		RunE: c.statementList,
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	c.addComputePoolFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.Flags().String("status", "", "Filter the results by statement status.")
	pcmd.RegisterFlagCompletionFunc(cmd, "status", func(*cobra.Command, []string) []string {
		return allowedStatuses
	})

	return cmd
}

func (c *command) statementList(cmd *cobra.Command, _ []string) error {
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
	status = strings.ToLower(status)

	if status != "" && !slices.Contains(allowedStatuses, status) {
		log.CliLogger.Warnf(`Invalid status "%s". Valid statuses are %s.`, status, utils.ArrayToCommaDelimitedString(allowedStatuses, "and"))
	}

	statements, err := client.ListStatements(environmentId, c.Context.GetCurrentOrganization(), c.Context.GetCurrentFlinkComputePool())
	if err != nil {
		return err
	}

	if status != "" {
		statements = lo.Filter(statements, func(statement v1beta1.SqlV1beta1Statement, _ int) bool {
			return strings.ToLower(statement.Status.GetPhase()) == status
		})
	}

	list := output.NewList(cmd)
	for _, statement := range statements {
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
