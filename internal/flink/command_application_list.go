package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the Flink environment.")
	cmd.Flags().String("filter", "", `Filter the applications with a CMF filter expression, for example "name=my-app*,state=RUNNING". Terms are comma-separated "key=value" pairs; "name" accepts a trailing "*" wildcard, and "state" values are case-insensitive.`)
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

	filter, err := cmd.Flags().GetString("filter")
	if err != nil {
		return err
	}

	pageSize, err := getPageSize(cmd)
	if err != nil {
		return err
	}

	client, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applications, err := client.ListApplications(c.createContext(), environment, normalizeApplicationFilter(filter), pageSize)
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

// normalizeApplicationFilter case-folds the value of any "state" term in a CMF applications
// filter expression to upper case. Flink job states are an upper-case enum and CMF matches
// them case-sensitively, so this lets users write "state=running" instead of "state=RUNNING".
// Every other term is passed through verbatim: "name" values in particular are case-sensitive
// (Kubernetes resource names), so they must not be folded. The grammar is the comma-separated
// "key=value" syntax of the CMF applications list "filter" query parameter.
func normalizeApplicationFilter(filter string) string {
	if filter == "" {
		return ""
	}

	terms := strings.Split(filter, ",")
	for i, term := range terms {
		key, value, found := strings.Cut(term, "=")
		if found && strings.EqualFold(strings.TrimSpace(key), "state") {
			terms[i] = "state=" + strings.ToUpper(strings.TrimSpace(value))
		}
	}
	return strings.Join(terms, ",")
}
