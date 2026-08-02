package flink

import (
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

// allowedApplicationStatuses lists the Flink job states recognized by the CMF applications
// "state=" filter, per the cmf-sdk-go GetApplications filter documentation. Unknown values are
// still forwarded (the server returns no matches rather than erroring); this list only drives
// the advisory --status warning.
var allowedApplicationStatuses = []string{"RUNNING", "FINISHED", "FAILED", "CANCELED", "RECONCILING", "COMPLETED", "UNKNOWN"}

func (c *command) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("name", "", `Filter the Flink applications by name. Supports wildcards, for example "my-app*".`)
	cmd.Flags().String("status", "", "Filter the Flink applications by status.")
	addPageSizeFlag(cmd)
	addCmfFlagSet(cmd)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))

	return cmd
}

func (c *command) applicationList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	status, err := cmd.Flags().GetString("status")
	if err != nil {
		return err
	}
	if status != "" {
		status = strings.ToUpper(status)
		if !slices.Contains(allowedApplicationStatuses, status) {
			output.ErrPrintf(c.Config.EnableColor, "[WARN] Invalid status %q. Valid statuses are %s.\n", status, utils.ArrayToCommaDelimitedString(allowedApplicationStatuses, "and"))
		}
	}

	pageSize, err := getPageSize(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applications, err := client.ListApplications(c.createContext(), environment, buildApplicationFilter(name, status), pageSize)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			status := app.GetStatus()
			rawJobStatus, ok := status["jobStatus"]
			if !ok {
				return fmt.Errorf("job status not found in flink job status")
			}
			jobStatus, ok := rawJobStatus.(map[string]interface{})
			if !ok {
				return fmt.Errorf("jobStatus has unexpected type")
			}
			envInApp, ok := app.Spec["environment"].(string)
			if !ok {
				envInApp = environment
			}
			list.Add(&flinkApplicationSummaryOut{
				Name:        app.Metadata["name"].(string),
				Environment: envInApp,
				JobName:     jobStatus["jobName"].(string),
				JobStatus:   jobStatus["state"].(string),
			})
		}
		return list.Print()
	}

	localApps := make([]LocalFlinkApplication, 0, len(applications))

	for _, sdkApp := range applications {
		localApps = append(localApps, convertSdkApplicationToLocalApplication(sdkApp))
	}

	return output.SerializedOutput(cmd, localApps)
}

// buildApplicationFilter composes the CMF applications "filter" query from the user-facing
// --name and --status flags. The grammar (comma-separated "key=value" expressions, "name="
// with an optional "*" suffix wildcard, and "state=" for status) follows the CMF applications
// list API. Values are not escaped: Kubernetes application names and Flink states cannot
// contain "," or "=", so no ambiguity arises. The caller normalizes and warns about status.
func buildApplicationFilter(name, status string) string {
	filters := make([]string, 0, 2)
	if name != "" {
		filters = append(filters, fmt.Sprintf("name=%s", name))
	}
	if status != "" {
		filters = append(filters, fmt.Sprintf("state=%s", status))
	}
	return strings.Join(filters, ",")
}
