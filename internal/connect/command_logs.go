package connect

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type logsCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type logEntryOut struct {
	Timestamp string `human:"Timestamp" serialized:"timestamp"`
	Level     string `human:"Level" serialized:"level"`
	TaskId    string `human:"Task ID" serialized:"task_id"`
	Message   string `human:"Message" serialized:"message"`
	Exception string `human:"Exception" serialized:"exception"`
}

func newLogsCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "Manage logs for connectors.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Query connector logs with log level ERROR between the provided time window:",
				Code: `confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"`,
			},
			examples.Example{
				Text: "Query connector logs with log level ERROR and WARN between the provided time window:",
				Code: `confluent connect logs lcc-123456 --level "ERROR|WARN" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"`,
			},
			examples.Example{
				Text: "Query subsequent pages of connector logs for the same query by executing the command with next flag until \"No logs found for the current query\" is printed to the console:",
				Code: `confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --next`,
			},
			examples.Example{
				Text: "Query connector logs with log level ERROR and containing \"example error\" in logs between the provided time window, and store in file:",
				Code: `confluent connect logs lcc-123456 --level "ERROR" --search-text "example error" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --output-file errors.json`,
			},
			examples.Example{
				Text: "Query connector logs with log level ERROR and matching regex \"exa*\" in logs between the provided time window, and store in file:",
				Code: `confluent connect logs lcc-123456 --level "ERROR" --search-text "exa*" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --output-file errors.json`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &logsCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.queryLogs
	cmd.Flags().String("start-time", "", "Start time for log query in ISO 8601 (https://en.wikipedia.org/wiki/ISO_8601) UTC datetime format (e.g., 2025-02-01T00:00:00Z).")
	cmd.Flags().String("end-time", "", "End time for log query in ISO 8601 (https://en.wikipedia.org/wiki/ISO_8601) UTC datetime format (e.g., 2025-02-01T23:59:59Z).")
	cmd.Flags().String("level", "ERROR", "Log level filter (INFO, WARN, ERROR). Defaults to ERROR. Use '|' to specify multiple levels (e.g., ERROR|WARN).")
	cmd.Flags().String("search-text", "", "Search text within logs.")
	cmd.Flags().String("output-file", "", "Output file path to append connector logs.")
	cmd.Flags().Bool("next", false, "Whether to fetch next page of logs after the next execution of the command.")

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("start-time"))
	cobra.CheckErr(cmd.MarkFlagRequired("end-time"))

	return cmd
}

func (c *logsCommand) queryLogs(cmd *cobra.Command, args []string) error {
	connectorId := args[0]
	if connectorId == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}

	startTime, err := cmd.Flags().GetString("start-time")
	if err != nil {
		return err
	}

	endTime, err := cmd.Flags().GetString("end-time")
	if err != nil {
		return err
	}

	level, err := cmd.Flags().GetString("level")
	if err != nil {
		return err
	}
	levels := strings.Split(level, "|")
	for _, l := range levels {
		if l != "INFO" && l != "WARN" && l != "ERROR" {
			return fmt.Errorf("invalid log level: %s", l)
		}
	}

	searchText, err := cmd.Flags().GetString("search-text")
	if err != nil {
		return err
	}

	outputFile, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return err
	}

	next, err := cmd.Flags().GetBool("next")
	if err != nil {
		return err
	}

	if err := validateTimeFormat(startTime); err != nil {
		return fmt.Errorf("invalid start-time format: %w", err)
	}

	if err := validateTimeFormat(endTime); err != nil {
		return fmt.Errorf("invalid end-time format: %w", err)
	}

	if endTime < startTime {
		return fmt.Errorf("end-time must be greater than start-time")
	}

	startTimeParsed, _ := time.Parse(time.RFC3339, startTime)
	now := time.Now()
	maxAge := 72 * time.Hour
	if startTimeParsed.Before(now.Add(-maxAge)) {
		return fmt.Errorf("start-time cannot be older than 72 hours")
	}

	currentLogQuery := &config.ConnectLogsQueryState{
		StartTime:   startTime,
		EndTime:     endTime,
		Level:       level,
		SearchText:  searchText,
		ConnectorId: connectorId,
		PageToken:   "",
	}

	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return fmt.Errorf("Kafka cluster information not found: %w\nPlease ensure you have set a cluster context with 'confluent kafka cluster use <cluster-id>' or specify --cluster flag", err)
	}
	kafkaClusterId := kafkaCluster.GetId()

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return fmt.Errorf("Environment ID not found: %w\nPlease ensure you have set an environment context with 'confluent environment use <env-id>' or specify --environment flag", err)
	}

	connectorName, err := c.getConnectorName(connectorId, environmentId, kafkaClusterId)
	if connectorName == "" {
		return fmt.Errorf("Connector not found: %w", err)
	}

	lastQueryPageToken, err := c.getPageTokenFromStoredQuery(next, currentLogQuery)
	if err != nil {
		return nil
	}

	crn := fmt.Sprintf("crn://confluent.cloud/organization=%s/environment=%s/cloud-cluster=%s/connector=%s",
		c.Context.GetCurrentOrganization(),
		environmentId,
		kafkaClusterId,
		connectorName,
	)

	logs, err := c.V2Client.SearchConnectorLogs(crn, connectorId, startTime, endTime, levels, searchText, lastQueryPageToken)
	if err != nil {
		return fmt.Errorf("failed to query connector logs: %w", err)
	}

	err = c.storeQueryInContext(logs, currentLogQuery)
	if err != nil {
		return err
	}

	if outputFile != "" {
		return writeLogsToFile(outputFile, logs)
	}

	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, logs.Data)
	}

	return printHumanLogs(cmd, logs, connectorId)
}

func (c *logsCommand) getPageTokenFromStoredQuery(next bool, currentLogQuery *config.ConnectLogsQueryState) (string, error) {
	lastLogQuery := c.Context.GetConnectLogsQueryState()
	var lastQueryPageToken string
	if next {
		if lastLogQuery != nil && (lastLogQuery.StartTime == currentLogQuery.StartTime &&
			lastLogQuery.EndTime == currentLogQuery.EndTime &&
			lastLogQuery.Level == currentLogQuery.Level &&
			lastLogQuery.SearchText == currentLogQuery.SearchText &&
			lastLogQuery.ConnectorId == currentLogQuery.ConnectorId) {
			lastQueryPageToken = lastLogQuery.PageToken
			if lastQueryPageToken == "" {
				output.Printf(false, "No logs found for the current query\n")
				return "", fmt.Errorf("No logs found for the current query")
			}
		} else {
			lastQueryPageToken = ""
		}
	} else {
		lastQueryPageToken = ""
	}
	return lastQueryPageToken, nil
}

func (c *logsCommand) storeQueryInContext(logs *ccloudv2.LoggingSearchResponse, currentLogQuery *config.ConnectLogsQueryState) error {
	if logs.Metadata != nil {
		pageToken, err := extractPageToken(logs.Metadata.Next)
		currentLogQuery.SetPageToken(pageToken)
		if err != nil {
			return fmt.Errorf("failed to extract page token: %w", err)
		}
	} else {
		currentLogQuery.SetPageToken("")
	}

	err := c.Context.SetConnectLogsQueryState(currentLogQuery)
	if err != nil {
		return fmt.Errorf("failed to set connect logs query state: %w", err)
	}
	if err := c.Config.Save(); err != nil {
		return err
	}
	return nil
}

func (c *logsCommand) getConnectorName(connectorId, environmentId, kafkaClusterId string) (string, error) {
	connector, err := c.V2Client.GetConnectorExpansionById(connectorId, environmentId, kafkaClusterId)
	if err != nil {
		return "", err
	}
	connectorInfo := connector.GetInfo()
	return connectorInfo.GetName(), nil
}

func writeLogsToFile(outputFile string, logs *ccloudv2.LoggingSearchResponse) error {
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", outputFile, err)
	}
	defer file.Close()

	for _, log := range logs.Data {
		logEntry := &logEntryOut{
			Timestamp: log.Timestamp,
			Level:     log.Level,
			TaskId:    log.TaskId,
			Message:   log.Message,
		}
		if log.Exception != nil {
			logEntry.Exception = log.Exception.Stacktrace
		}
		data, err := json.Marshal(logEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry to JSON: %w", err)
		}

		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write log entry to file %s: %w", outputFile, err)
		}

		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline to file %s: %w", outputFile, err)
		}
	}

	output.Printf(false, "Appended %d log entries to file: %s\n", len(logs.Data), outputFile)
	return nil
}

func validateTimeFormat(timeStr string) error {
	pattern := `^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`
	match, err := regexp.MatchString(pattern, timeStr)
	if !match || err != nil {
		return fmt.Errorf("must be formatted as: YYYY-MM-DDTHH:MM:SSZ")
	}
	return nil
}

func extractPageToken(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	queryParams := parsedURL.Query()
	pageToken := queryParams.Get("page_token")

	return pageToken, nil
}

func printHumanLogs(cmd *cobra.Command, logs *ccloudv2.LoggingSearchResponse, connectorId string) error {
	list := output.NewList(cmd)
	for _, log := range logs.Data {
		logOut := &logEntryOut{
			Timestamp: log.Timestamp,
			Level:     log.Level,
			TaskId:    log.TaskId,
			Message:   log.Message,
		}
		if log.Exception != nil {
			logOut.Exception = log.Exception.Stacktrace
		}
		list.Add(logOut)
	}

	if len(logs.Data) == 0 {
		output.Println(false, "No logs found for the current query")
		return nil
	}

	output.Printf(false, "Found %d log entries for connector %s:\n\n", len(logs.Data), connectorId)
	return list.Print()
}
