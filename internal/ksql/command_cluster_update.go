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

// Valid CSU sizes that customers may target via self-serve cluster update.
// Mirrors the server-side authoritative list in cc-control-plane-ksql:
// internal/service/update_ksql_cluster_resize.go::validCSUSizes.
// Values 1, 2 are legacy and not user-selectable. Values above maxSelfServeCSU
// (the largest entry in this slice) still require a support ticket.
//
//nolint:gochecknoglobals
var validCsuSizes = []int32{4, 8, 12, 16, 20, 24, 28}

// maxSelfServeCSU is derived from validCsuSizes so the support-ticket
// threshold and the "Valid values" listing stay in lockstep — if the
// validCsuSizes slice is extended, the threshold moves with it.
//
//nolint:gochecknoglobals
var maxSelfServeCSU = func() int32 {
	max := int32(0)
	for _, v := range validCsuSizes {
		if v > max {
			max = v
		}
	}
	return max
}()

func csuSupportTicketMessage() string {
	return fmt.Sprintf(
		"CSU values above %d require a support ticket. "+
			"Please contact Confluent Support to request a larger cluster size.",
		maxSelfServeCSU)
}

func (c *ksqlCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update a ksqlDB cluster.",
		Long:              buildUpdateLongDescription(),
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		// Hidden while the SDK call is shimmed (see Client.UpdateKsqlCluster
		// in pkg/ccloudv2/ksql.go). Once the SDK is regenerated from cc-api
		// PR #2507 and the shim is replaced with the real call, drop Hidden.
		Hidden: true,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Resize ksqlDB cluster "lksqlc-12345" to 8 CSUs.`,
				Code: "confluent ksql cluster update lksqlc-12345 --csu 8",
			},
		),
	}

	cmd.Flags().Int32("csu", 0, fmt.Sprintf(
		"Target number of CSUs for the cluster. Valid values: %s.",
		formatCsuList(validCsuSizes)))
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("csu"))

	return cmd
}

func buildUpdateLongDescription() string {
	return fmt.Sprintf(
		"Update an existing ksqlDB cluster. Currently only the CSU count may be modified, "+
			"and only to larger sizes (shrink is not supported).\n\n"+
			"Valid CSU values are %s. Larger sizes require a support ticket. "+
			"The cluster will undergo a rolling restart to apply the new size; "+
			"the command returns once the resize has been accepted by the control plane.",
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

	// Pre-check current CSU so we can short-circuit a no-op locally before
	// issuing the PATCH. The server-side validator also rejects no-op resizes
	// with 400 ("new CSU size is the same as old CSU size, no-op"), but a
	// client-side check produces a clearer message and avoids a wasted API
	// round trip. Note: shrink is not supported server-side either.
	current, err := c.V2Client.DescribeKsqlCluster(clusterId, environmentId)
	if err != nil {
		return errors.CatchKSQLNotFoundError(err, clusterId)
	}
	currentCsu := current.Spec.GetCsu()
	if currentCsu == csu {
		return fmt.Errorf("ksqlDB cluster %q is already at %d CSUs; no change requested",
			clusterId, csu)
	}
	if csu < currentCsu {
		return fmt.Errorf("ksqlDB cluster %q is currently %d CSUs; shrinking is not supported "+
			"(target %d < current %d)", clusterId, currentCsu, csu, currentCsu)
	}

	cluster, err := c.V2Client.UpdateKsqlCluster(clusterId, environmentId, csu)
	if err != nil {
		return err
	}

	// Print the rolling-restart notice only AFTER the PATCH was accepted —
	// otherwise a failed call (e.g., a 4xx from the server) would leave the
	// customer with a misleading "Resizing…" message even though no resize
	// is happening.
	output.ErrPrintf(c.Config.EnableColor,
		"Resizing ksqlDB cluster %q from %d to %d CSUs. A rolling restart will be "+
			"performed asynchronously; the cluster will continue serving queries during the resize.\n",
		clusterId, currentCsu, csu)

	table := output.NewTable(cmd)
	table.Add(c.formatClusterForDisplayAndList(&cluster))
	return table.Print()
}

// validateCsuForUpdate returns nil if csu is in validCsuSizes, and a
// customer-safe error otherwise. The server-side check in
// cc-control-plane-ksql is authoritative; this client-side validation exists
// to fail fast with a clearer message before issuing the API call.
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
