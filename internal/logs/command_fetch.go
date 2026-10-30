package logs

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/spf13/cobra"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type fetchCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newFetchCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fetch [resource-id]",
		Short: "Fetch logs from a log source.",
		Long:  "Fetch logs from a log source, for a single resource, or for a whole environment or Organization when the log source supports it. Run `confluent logs source list` to see the available log sources, and `confluent logs source describe <source>` for the search fields and returned fields of one.",
		Args:  cobra.MaximumNArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Fetch cluster health events for one Kafka cluster from the last hour:",
				Code: "confluent logs fetch lkc-123456 --source cluster-health --since 1h",
			},
			examples.Example{
				Text: "Fetch only ERROR events of a given type from the last 3 days, newest 100 entries:",
				Code: "confluent logs fetch lkc-123456 --source cluster-health --since 3d --search level=ERROR,event_type=SERVICE_ISSUE_DETECTED --tail 100",
			},
			examples.Example{
				Text: "Fetch cluster health events for a whole environment:",
				Code: "confluent logs fetch --source cluster-health --environment env-123456 --since 24h",
			},
			examples.Example{
				Text: "Fetch cluster health events for the whole organization:",
				Code: "confluent logs fetch --source cluster-health --since 24h",
			},
			examples.Example{
				Text: "Fetch ERROR and WARN entries containing the text \"example error\" and append them to a file as JSON lines:",
				Code: `confluent logs fetch lcc-123456 --source connector --since 24h --search level=ERROR,level=WARN --search-text "example error" --output-file errors.jsonl`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &fetchCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.fetch

	cmd.Flags().String("source", "", `The log source to query, for example "cluster-health". Run "confluent logs source list" for the available log sources.`)
	cmd.Flags().String("since", "", "Relative time window ending now, for example 1h, 24h or 7d. Cannot be combined with --from or --to.")
	cmd.Flags().String("from", "", "Window start, in ISO 8601 UTC format (for example 2026-08-20T00:00:00Z).")
	cmd.Flags().String("to", "", "Window end, in ISO 8601 UTC format. Defaults to now when --from is set.")
	cmd.Flags().StringSlice("search", nil, `A comma-separated list of "field=value" search filters, validated against the log source. Repeating a field means OR (for example level=ERROR,level=WARN).`)
	cmd.Flags().String("search-text", "", "Full-text search within log messages, where the log source supports it.")
	cmd.Flags().Int("tail", 0, fmt.Sprintf("Return only the most recent N entries (maximum %d).", maxTail))
	cmd.Flags().String("output-file", "", "Append the results to this file as JSON lines.")

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("source"))
	cmd.MarkFlagsMutuallyExclusive("since", "from")
	cmd.MarkFlagsMutuallyExclusive("since", "to")

	return cmd
}

func (c *fetchCommand) fetch(cmd *cobra.Command, args []string) error {
	resourceId := ""
	if len(args) > 0 {
		resourceId = args[0]
		if resourceId == "" {
			return fmt.Errorf("resource ID cannot be empty")
		}
	}

	sourceName, err := cmd.Flags().GetString("source")
	if err != nil {
		return err
	}
	since, err := cmd.Flags().GetString("since")
	if err != nil {
		return err
	}
	from, err := cmd.Flags().GetString("from")
	if err != nil {
		return err
	}
	to, err := cmd.Flags().GetString("to")
	if err != nil {
		return err
	}
	pairs, err := cmd.Flags().GetStringSlice("search")
	if err != nil {
		return err
	}
	searchText, err := cmd.Flags().GetString("search-text")
	if err != nil {
		return err
	}
	tail, err := cmd.Flags().GetInt("tail")
	if err != nil {
		return err
	}
	outputFile, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return err
	}

	if tail < 0 || tail > maxTail {
		return fmt.Errorf(tailOutOfRangeErrorMsg, maxTail)
	}

	startTime, endTime, err := resolveTimeWindow(since, from, to, time.Now())
	if err != nil {
		return err
	}

	search, err := buildSearch(pairs, searchText)
	if err != nil {
		return err
	}

	source, err := c.V2Client.DescribeLogSource(sourceName)
	if err != nil {
		return err
	}

	if err := validateSearch(source, search); err != nil {
		return err
	}
	if err := validateLookback(source, since); err != nil {
		return err
	}
	if err := validateResourceId(source, resourceId); err != nil {
		return err
	}

	scope := scopeFor(resourceId, cmd.Flags().Changed("environment"))
	if err := validateScope(source, scope, resourceId); err != nil {
		return err
	}

	crn, err := c.buildCRN(source, scope, resourceId)
	if err != nil {
		return err
	}

	pageSize := ccloudv2.LogsPageSize
	if tail > 0 {
		pageSize = tail
	}

	logs, err := c.V2Client.SearchLogs(ccloudv2.LogsSearchRequest{
		CRN:       crn,
		Source:    source.Name,
		Search:    search,
		Sort:      "desc",
		StartTime: startTime,
		EndTime:   endTime,
	}, pageSize, "")
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	if outputFile != "" {
		return writeLogsToFile(outputFile, logs.Data)
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, logs.Data)
	}

	return printHumanLogs(cmd, source, resourceId, logs.Data)
}

// buildCRN resolves the CRN element values from the command's inputs and the active context.
func (c *fetchCommand) buildCRN(source *loggingv1.LoggingV1Source, scope, resourceId string) (string, error) {
	values := crnValues{Organization: c.Context.GetCurrentOrganization()}

	if scope != scopeOrganization {
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return "", errors.NewWrapErrorWithSuggestions(err, environmentNotFoundErrorMsg, environmentNotFoundSuggestions)
		}
		values.Environment = environmentId
	}

	if scope == scopeResource {
		values.Leaf = resourceId

		switch {
		case source.ResourceType == "kafka":
			// A Kafka cluster's ID is both its cloud-cluster and its leaf element.
			values.Cluster = resourceId
		case containsElement(source.CrnElements, "cloud-cluster"):
			cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
			if err != nil {
				return "", errors.NewWrapErrorWithSuggestions(err, kafkaClusterNotFoundErrorMsg, kafkaClusterNotFoundSuggestions)
			}
			values.Cluster = cluster.GetId()

			if source.ResourceType == "connector" {
				connector, err := c.V2Client.GetConnectorExpansionById(resourceId, values.Environment, values.Cluster)
				if err != nil {
					return "", fmt.Errorf("connector %q not found: %w", resourceId, err)
				}
				info := connector.GetInfo()
				values.Leaf = info.GetName()
			}
		}
	}

	return buildCRN(source, scope, values)
}

func containsElement(elements []string, element string) bool {
	for _, e := range elements {
		if e == element {
			return true
		}
	}
	return false
}

func writeLogsToFile(path string, records []map[string]any) error {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	for _, record := range records {
		line, err := json.Marshal(record)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry: %w", err)
		}
		if _, err := file.Write(append(line, '\n')); err != nil {
			return fmt.Errorf("failed to write to file %s: %w", path, err)
		}
	}

	output.Printf(false, "Appended %d log entries to file: %s\n", len(records), path)
	return nil
}

// printHumanLogs prints the default columns the log source declares. The columns are only known at runtime, so the
// row type is built with reflection and rendered by the standard list printer.
func printHumanLogs(cmd *cobra.Command, source *loggingv1.LoggingV1Source, resourceId string, records []map[string]any) error {
	if len(records) == 0 {
		output.Println(false, "No logs found for the current query.")
		return nil
	}

	var columns []loggingv1.LoggingV1ReturnedFieldDescriptor
	for _, field := range source.ReturnedFields {
		if field.DefaultColumn {
			columns = append(columns, field)
		}
	}
	if len(columns) == 0 {
		return output.SerializedOutput(cmd, records)
	}

	fields := make([]reflect.StructField, len(columns))
	for i, column := range columns {
		fields[i] = reflect.StructField{
			Name: fmt.Sprintf("Column%d", i),
			Type: reflect.TypeOf(""),
			Tag:  reflect.StructTag(fmt.Sprintf(`human:%q serialized:%q`, columnHeader(column.Name), column.Name)),
		}
	}
	rowType := reflect.StructOf(fields)

	list := output.NewList(cmd)
	list.Sort(false) // keep the server's order (newest first)
	for _, record := range records {
		row := reflect.New(rowType)
		for i, column := range columns {
			row.Elem().Field(i).SetString(recordValue(record, column.Name))
		}
		list.Add(row.Interface())
	}

	noun := "entries"
	if len(records) == 1 {
		noun = "entry"
	}
	target := source.Name
	if resourceId != "" {
		target += " on " + resourceId
	}
	output.Printf(false, "Found %d log %s for %s:\n\n", len(records), noun, target)

	return list.Print()
}

// recordValue reads a field from a record, allowing for dotted names such as "exception.stacktrace" whether the API
// returns them as a flat key or as a nested object.
func recordValue(record map[string]any, name string) string {
	if value, ok := record[name]; ok {
		return formatValue(value)
	}

	var current any = record
	for _, part := range strings.Split(name, ".") {
		m, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		if current, ok = m[part]; !ok {
			return ""
		}
	}
	return formatValue(current)
}

func formatValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case float64:
		return fmt.Sprintf("%v", v)
	default:
		out, err := json.Marshal(v)
		if err != nil {
			return fmt.Sprint(v)
		}
		return string(out)
	}
}

// columnHeader turns a field name such as "event_type" into a column header such as "Event Type".
func columnHeader(name string) string {
	words := strings.FieldsFunc(name, func(r rune) bool { return r == '_' || r == '.' || r == '-' })
	for i, word := range words {
		if strings.EqualFold(word, "id") {
			words[i] = "ID"
			continue
		}
		words[i] = strings.ToUpper(word[:1]) + word[1:]
	}
	return strings.Join(words, " ")
}
