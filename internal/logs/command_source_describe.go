package logs

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type sourceDescribeCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newSourceDescribeCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <source>",
		Short: "Describe a log source.",
		Long:  "Describe a log source: the scopes it supports, the search fields `confluent logs fetch --search` accepts for it, and the fields each log record carries. Results are always fetched fresh.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "cluster-health" log source:`,
				Code: "confluent logs source describe cluster-health",
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &sourceDescribeCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.describe
	cmd.ValidArgsFunction = pcmd.NewValidArgsFunction(c.validArgs)

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *sourceDescribeCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return completeSources(c.AuthenticatedCLICommand, cmd, args)
}

func (c *sourceDescribeCommand) describe(cmd *cobra.Command, args []string) error {
	source, err := c.V2Client.DescribeLogSource(args[0])
	if err != nil {
		return err
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutputFromJsonTags(cmd, source)
	}

	output.Print(false, renderSource(source))
	return nil
}

// renderSource lays out everything a user needs to build a query and read its results. Fields shown in the default
// human output of `fetch` are marked with a star and listed first.
func renderSource(source *loggingv1.LoggingV1Source) string {
	var b strings.Builder

	resource := source.ResourceType
	var details []string
	if source.GetIdPrefix() != "" {
		details = append(details, source.GetIdPrefix()+"- IDs")
	}
	if len(source.CrnElements) > 0 {
		details = append(details, "crn: "+strings.Join(source.CrnElements, "/"))
	}
	if len(details) > 0 {
		resource += " (" + strings.Join(details, "; ") + ")"
	}

	fullText := "no"
	if source.FullTextSearchSupported {
		fullText = "yes"
		if len(source.GetFullTextSearchFields()) > 0 {
			fullText += " (--search-text over: " + strings.Join(source.GetFullTextSearchFields(), ", ") + ")"
		}
	}

	const labelWidth = len("Full-text search:") + 2
	line := func(label, value string) {
		fmt.Fprintf(&b, "%-*s%s\n", labelWidth, label+":", value)
	}
	line("Log source", source.Name)
	line("Display name", source.DisplayName)
	line("Resource", resource)
	line("Scopes", formatScopes(source.AllowedScopes))
	line("Max lookback", fmt.Sprintf("%dd", source.MaxLookbackDays))
	line("Full-text search", fullText)

	if len(source.SearchFields) > 0 {
		nameWidth, typeWidth := 0, 0
		for _, field := range source.SearchFields {
			nameWidth = max(nameWidth, len(field.Name))
			typeWidth = max(typeWidth, len(field.Type))
		}

		b.WriteString("\nSearch fields (--search field=value):\n")
		for _, field := range source.SearchFields {
			row := fmt.Sprintf("  %-*s  %-*s", nameWidth, field.Name, typeWidth, field.Type)
			if len(field.GetAllowedValues()) > 0 {
				row += "  " + strings.Join(field.GetAllowedValues(), ", ")
			}
			b.WriteString(strings.TrimRight(row, " ") + "\n")
		}
	}

	if len(source.ReturnedFields) > 0 {
		var fields []loggingv1.LoggingV1ReturnedFieldDescriptor
		for _, defaults := range []bool{true, false} {
			for _, field := range source.ReturnedFields {
				if field.DefaultColumn == defaults {
					fields = append(fields, field)
				}
			}
		}

		nameWidth, typeWidth := 0, 0
		for _, field := range fields {
			nameWidth = max(nameWidth, len(field.Name))
			typeWidth = max(typeWidth, len(field.Type))
		}

		b.WriteString("\nReturned fields (each log record):\n")
		for _, field := range fields {
			star := " "
			if field.DefaultColumn {
				star = "*"
			}
			row := fmt.Sprintf("  %-*s  %-*s  %s  %s", nameWidth, field.Name, typeWidth, field.Type, star, field.GetDescription())
			b.WriteString(strings.TrimRight(row, " ") + "\n")
		}
		b.WriteString("\n  * = default column in human output\n")
	}

	return b.String()
}
