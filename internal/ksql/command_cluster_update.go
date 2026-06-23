package ksql

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

// validCsuSizes mirrors cc-control-plane-ksql's authoritative list.
//
//nolint:gochecknoglobals
var validCsuSizes = []int32{4, 8, 12, 16, 20, 24, 28}

//nolint:gochecknoglobals
var maxSelfServeCSU = func() int32 {
	maxCsu := int32(0)
	for _, v := range validCsuSizes {
		if v > maxCsu {
			maxCsu = v
		}
	}
	return maxCsu
}()

func csuSupportTicketMessage() string {
	return fmt.Sprintf(
		"CSU values above %d require a support ticket. "+
			"Please contact Confluent Support to request a larger cluster size.",
		maxSelfServeCSU)
}

// buildUpdateExamples returns the customer-facing help examples for the
// update command — extracted so it's directly unit-testable.
func buildUpdateExamples() string {
	return examples.BuildExampleString(
		examples.Example{
			Text: `Expand ksqlDB cluster "lksqlc-12345" to 8 CSUs.`,
			Code: "confluent ksql cluster update lksqlc-12345 --csu 8",
		},
		examples.Example{
			Text: `Shrink ksqlDB cluster "lksqlc-12345" to 4 CSUs.`,
			Code: "confluent ksql cluster update lksqlc-12345 --csu 4",
		},
	)
}

// buildCsuFlagUsage returns the help text for the --csu flag — extracted
// so it's directly unit-testable.
func buildCsuFlagUsage() string {
	return fmt.Sprintf("Target number of CSUs for the cluster. Valid values: %s.",
		formatCsuList(validCsuSizes))
}

func (c *ksqlCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a ksqlDB cluster.",
		Long:              buildUpdateLongDescription(),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Hidden:            true, // until cc-api #2507 merges + public SDK regenerates
		Example:           buildUpdateExamples(),
	}

	cmd.Flags().Int32("csu", 0, buildCsuFlagUsage())
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("csu"))

	return cmd
}

func buildUpdateLongDescription() string {
	return fmt.Sprintf(
		"Update an existing ksqlDB cluster. Currently only the CSU count may be modified. "+
			"Both expansion (increase) and shrink (decrease) are supported.\n\n"+
			"Valid CSU values are %s. Larger sizes require a support ticket. "+
			"The cluster will undergo a rolling restart to apply the new size; "+
			"the command returns once the resize has been accepted by the control plane. "+
			"Shrink requests are precondition-checked against the cluster's running "+
			"persistent-query count and refused if the new size cannot host them; "+
			"drop excess queries with `TERMINATE <query_id>` and retry.",
		formatCsuList(validCsuSizes))
}

func (c *ksqlCommand) update(cmd *cobra.Command, args []string) error {
	csu, err := cmd.Flags().GetInt32("csu")
	if err != nil {
		return err
	}
	if err := validateCsuForUpdate(csu); err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	clusterId := args[0]

	// Client-side no-op short-circuit; direction is server-arbitrated.
	current, err := c.V2Client.DescribeKsqlCluster(clusterId, environmentId)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, clusterId)
	}
	currentCsu := current.Spec.GetCsu()
	if currentCsu == csu {
		return fmt.Errorf("ksqlDB cluster %q is already at %d CSUs; no change requested",
			clusterId, csu)
	}

	cluster, err := c.V2Client.UpdateKsqlCluster(clusterId, environmentId, csu)
	if err != nil {
		return err
	}

	// Rolling-restart notice prints only AFTER the PATCH was accepted.
	output.ErrPrintf(c.Config.EnableColor,
		"Resizing ksqlDB cluster %q from %d to %d CSUs. A rolling restart will be "+
			"performed asynchronously; the cluster will continue serving queries during the resize.\n",
		clusterId, currentCsu, csu)

	table := output.NewTable(cmd)
	table.Add(c.formatClusterForDisplayAndList(&cluster))
	return table.Print()
}

// validateCsuForUpdate fail-fast checks before issuing the API call.
func validateCsuForUpdate(csu int32) error {
	if csu > maxSelfServeCSU {
		return fmt.Errorf("%d CSUs: %s", csu, csuSupportTicketMessage())
	}
	for _, valid := range validCsuSizes {
		if csu == valid {
			return nil
		}
	}
	return fmt.Errorf("%d is not a valid CSU size for cluster update. Valid sizes are %s",
		csu, formatCsuList(validCsuSizes))
}

func formatCsuList(sizes []int32) string {
	sorted := append([]int32(nil), sizes...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	out := ""
	for i, s := range sorted {
		if i > 0 {
			out += ", "
		}
		out += fmt.Sprintf("%d", s)
	}
	return out
}
