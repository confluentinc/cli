package logs

import (
	"strings"
	"time"

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
		Long:  "Describe a log source: the search fields it accepts and the fields each log record carries. Results are always fetched fresh.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the "cluster-health" log source.`,
				Code: "confluent logs source describe cluster-health",
			},
		),
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
	// The source is always fetched fresh, and it refreshes the cached entry other logs commands reuse.
	storeLogSource(c.AuthenticatedCLICommand, discoveryCacheKey(c.Context.GetPlatformServer(), c.Context.GetCurrentOrganization()), source, time.Now())

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutputFromJsonTags(cmd, source)
	}

	return printHumanSource(cmd, source)
}

type sourceDescribeOut struct {
	Name            string `human:"Name"`
	DisplayName     string `human:"Display Name"`
	Resource        string `human:"Resource"`
	IdPrefix        string `human:"ID Prefix"`
	CrnElements     string `human:"CRN Elements"`
	MaxLookbackDays int32  `human:"Max Lookback Days"`
	FullTextSearch  bool   `human:"Full-Text Search"`
	FullTextFields  string `human:"Full-Text Fields"`
}

type searchFieldOut struct {
	Name          string `human:"Name"`
	Type          string `human:"Type"`
	AllowedValues string `human:"Allowed Values"`
}

type returnedFieldOut struct {
	Name        string `human:"Name"`
	Type        string `human:"Type"`
	Default     bool   `human:"Default"`
	Description string `human:"Description"`
}

// printHumanSource prints everything a user needs to build a query and read its results: the log source, the search
// fields it accepts, and the fields each record carries. The fields shown in the default human output of `fetch` are
// marked as Default and listed first.
func printHumanSource(cmd *cobra.Command, source *loggingv1.LoggingV1Source) error {
	output.Println(false, "Log Source")
	table := output.NewTable(cmd)
	table.Add(newSourceDescribeOut(source))
	if err := table.Print(); err != nil {
		return err
	}

	if searchFields := newSearchFieldOuts(source); len(searchFields) > 0 {
		output.Println(false, "")
		output.Println(false, "Search Fields (--search field=value)")
		list := output.NewList(cmd)
		list.Sort(false) // keep the order the API returns
		for _, field := range searchFields {
			list.Add(field)
		}
		if err := list.Print(); err != nil {
			return err
		}
	}

	if returnedFields := newReturnedFieldOuts(source); len(returnedFields) > 0 {
		output.Println(false, "")
		output.Println(false, "Returned Fields (each log record)")
		list := output.NewList(cmd)
		list.Sort(false) // keep the default columns first
		for _, field := range returnedFields {
			list.Add(field)
		}
		return list.Print()
	}

	return nil
}

func newSourceDescribeOut(source *loggingv1.LoggingV1Source) *sourceDescribeOut {
	return &sourceDescribeOut{
		Name:            source.Name,
		DisplayName:     source.DisplayName,
		Resource:        source.ResourceType,
		IdPrefix:        source.GetIdPrefix(),
		CrnElements:     strings.Join(source.CrnElements, "/"),
		MaxLookbackDays: source.MaxLookbackDays,
		FullTextSearch:  source.FullTextSearchSupported,
		FullTextFields:  strings.Join(source.GetFullTextSearchFields(), ", "),
	}
}

func newSearchFieldOuts(source *loggingv1.LoggingV1Source) []*searchFieldOut {
	fields := make([]*searchFieldOut, 0, len(source.SearchFields))
	for _, field := range source.SearchFields {
		fields = append(fields, &searchFieldOut{
			Name:          field.Name,
			Type:          field.Type,
			AllowedValues: strings.Join(field.GetAllowedValues(), ", "),
		})
	}
	return fields
}

// newReturnedFieldOuts lists the default columns first, then the rest, each group in the order the API returns them.
func newReturnedFieldOuts(source *loggingv1.LoggingV1Source) []*returnedFieldOut {
	fields := make([]*returnedFieldOut, 0, len(source.ReturnedFields))
	for _, defaults := range []bool{true, false} {
		for _, field := range source.ReturnedFields {
			if field.DefaultColumn == defaults {
				fields = append(fields, &returnedFieldOut{
					Name:        field.Name,
					Type:        field.Type,
					Default:     field.DefaultColumn,
					Description: field.GetDescription(),
				})
			}
		}
	}
	return fields
}
