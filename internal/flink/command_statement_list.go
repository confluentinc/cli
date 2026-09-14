package flink

import (
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *statementCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink SQL statements.",
		Args:  cobra.NoArgs,
		Example: examples.BuildExampleString(examples.Example{
			Text: "List running statements.",
			Code: "confluent flink statement list --status running",
		}),
		RunE: c.list,
	}

	// Optional flags
	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
	pcmd.AddComputePoolFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("status", "", "Filter the results by statement status.")
	pcmd.RegisterFlagCompletionFunc(cmd, "status", func(*cobra.Command, []string) []string {
		return allowedStatuses
	})
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *statementCommand) list(cmd *cobra.Command, _ []string) error {
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

	computePoolId := c.Context.GetCurrentFlinkComputePool()
	if err := c.validateProvidedComputePool(environmentId, computePoolId); err != nil {
		return err
	}

	statements, err := client.ListStatements(environmentId, c.Context.GetCurrentOrganization(), computePoolId)
	if err != nil {
		return err
	}

	if status != "" {
		statements = lo.Filter(statements, func(statement flinkgatewayv1.SqlV1Statement, _ int) bool {
			return strings.ToLower(statement.Status.GetPhase()) == status
		})
	}

	list := output.NewList(cmd)
	for _, statement := range statements {
		list.Add(&statementOut{
			CreationDate:           statement.Metadata.GetCreatedAt(),
			Name:                   statement.GetName(),
			Statement:              statement.Spec.GetStatement(),
			ComputePool:            statement.Spec.GetComputePoolId(),
			Status:                 statement.Status.GetPhase(),
			StatusDetail:           statement.Status.GetDetail(),
			Warnings:               types.NewStatementWarnings(statement.Status.GetWarnings()),
			LatestOffsets:          statement.Status.GetLatestOffsets(),
			LatestOffsetsTimestamp: flinkgatewayv1.PtrTime(statement.Status.GetLatestOffsetsTimestamp()),
		})
	}
	return list.Print()
}

func (c *statementCommand) validateProvidedComputePool(environmentId, computePoolId string) error {
	if computePoolId == "" {
		return nil
	}

	computePool, err := c.V2Client.DescribeFlinkComputePool(computePoolId, environmentId)
	if err != nil {
		return err
	}

	computePoolCloudProvider := strings.ToLower(computePool.Spec.GetCloud())
	computePoolFlinkRegion := strings.ToLower(computePool.Spec.GetRegion())

	providedCloudProvider := strings.ToLower(c.Context.GetCurrentFlinkCloudProvider())
	providedFlinkRegion := strings.ToLower(c.Context.GetCurrentFlinkRegion())

	if computePoolCloudProvider != providedCloudProvider ||
		computePoolFlinkRegion != providedFlinkRegion {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf("Flink compute pool %q not found in %s %s", computePoolId, providedCloudProvider, providedFlinkRegion),
			fmt.Sprintf("Select a different compute pool, or provide the correct cloud/region pair with `confluent flink statement list --cloud %s --region %s --compute-pool %s --environment %s`",
				computePoolCloudProvider, computePoolFlinkRegion, computePoolId, environmentId))
	}

	return nil
}
