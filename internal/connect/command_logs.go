package connect

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
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
	TaskId    string `human:"Task ID,omitempty" serialized:"task_id,omitempty"`
	Message   string `human:"Message" serialized:"message"`
}

func newLogsCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs <id>",
		Short: "Query logs for connectors.",
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Query connector logs with log level ERROR between the provided time window:",
				Code: `confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"`,
			},
			examples.Example{
				Text: "Query next page of connector logs for the same query by running the command repeatedly until \"No more logs for the current query\" is printed to the console:",
				Code: `confluent connect logs lcc-123456 --level ERROR --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --next`,
			},
			examples.Example{
				Text: "Query all connector logs between the provided time window:",
				Code: `confluent connect logs lcc-123456 --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z"`,
			},
			examples.Example{
				Text: "Query connector logs with log level ERROR and containing \"example error\" in logs between the provided time window, and store in file:",
				Code: `confluent connect logs lcc-123456 --level "ERROR" --search-text "example error" --start-time "2025-02-01T00:00:00Z" --end-time "2025-02-01T23:59:59Z" --output-file errors.json`,
			},
		),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &logsCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.RunE = c.queryLogs
	cmd.Flags().String("start-time", "", "Start time for log query (e.g., 2025-02-01T00:00:00Z).")
	cmd.Flags().String("end-time", "", "End time for log query (e.g., 2025-02-01T23:59:59Z).")
	cmd.Flags().String("level", "ERROR", "Log level filter (INFO, WARN, ERROR). Defaults to ERROR.")
	cmd.Flags().String("search-text", "", "Search text within logs (optional).")
	cmd.Flags().String("output-file", "", "Output file path to append connector logs (optional).")
	cmd.Flags().Bool("next", false, "Whether to fetch next page of logs after the next execution of the command (optional).")

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("start-time"))
	cobra.CheckErr(cmd.MarkFlagRequired("end-time"))

	return cmd
}

func (c *logsCommand) queryLogs(cmd *cobra.Command, args []string) error {
	connectorId := args[0]

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

	currentLogQuery := &config.ConnectLogsQueryState{
		StartTime:   startTime,
		EndTime:     endTime,
		Level:       level,
		SearchText:  searchText,
		ConnectorId: connectorId,
		PageToken:   "",
	}
	// Validate time format
	if err := validateTimeFormat(startTime); err != nil {
		return fmt.Errorf("invalid start-time format: %w", err)
	}

	if err := validateTimeFormat(endTime); err != nil {
		return fmt.Errorf("invalid end-time format: %w", err)
	}

	kafkaCluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return fmt.Errorf("failed to get Kafka cluster information: %w\nPlease ensure you have set a cluster context with 'confluent kafka cluster use <cluster-id>' or specify --cluster flag", err)
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return fmt.Errorf("failed to get environment ID: %w\nPlease ensure you have set an environment context with 'confluent environment use <env-id>' or specify --environment flag", err)
	}

	// Validate connector ID format (basic validation)
	if connectorId == "" {
		return fmt.Errorf("connector ID cannot be empty")
	}

	// Query logs using the V2Client
	levels := []string{level}

	// Add debug information for verbose mode
	if verbosity, _ := cmd.Flags().GetCount("verbose"); verbosity >= 2 { // info level
		output.Printf(c.Config.EnableColor, "Making logs API request with:\n")
		output.Printf(c.Config.EnableColor, "  Connector ID: %s\n", connectorId)
		output.Printf(c.Config.EnableColor, "  Environment: %s\n", environmentId)
		output.Printf(c.Config.EnableColor, "  Cluster: %s\n", kafkaCluster.ID)
		output.Printf(c.Config.EnableColor, "  Time range: %s to %s\n", startTime, endTime)
		output.Printf(c.Config.EnableColor, "  Log levels: %v\n", levels)
		if searchText != "" {
			output.Printf(c.Config.EnableColor, "  Search text: %s\n", searchText)
		}
	}

	connector, err := c.V2Client.GetConnectorExpansionById(connectorId, environmentId, kafkaCluster.ID)
	if err != nil {
		return err
	}

	connectorName := connector.Info.GetName()
	lastLogQuery := c.Context.GetConnectLogsQueryState()
	var lastQueryPageToken string
	if next {
		if lastLogQuery != nil && (lastLogQuery.StartTime == startTime &&
			lastLogQuery.EndTime == endTime &&
			lastLogQuery.Level == level &&
			lastLogQuery.SearchText == searchText &&
			lastLogQuery.ConnectorId == connectorId) {
			lastQueryPageToken = lastLogQuery.PageToken
			if lastQueryPageToken == "" {
				output.Printf(c.Config.EnableColor, "No more logs for the current query\n")
				return nil
			}
		} else {
			lastQueryPageToken = ""
		}
	} else {
		lastQueryPageToken = ""
	}

	logs, err := c.V2Client.SearchConnectorLogs(environmentId, kafkaCluster.ID, connectorName, startTime, endTime, levels, searchText, 200, lastQueryPageToken)
	if err != nil {
		// Add context to the error
		return fmt.Errorf("failed to query connector logs: %w", err)
	}

	if logs.Metadata != nil {
		currentLogQuery.PageToken, err = extractPageToken(logs.Metadata.Next)
		if err != nil {
			return fmt.Errorf("failed to extract page token: %w", err)
		}
	} else {
		currentLogQuery.PageToken = ""
	}

	err = c.Context.SetConnectLogsQueryState(currentLogQuery)
	if err != nil {
		return fmt.Errorf("failed to set connect logs query state: %w", err)
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	// Handle output to file if specified
	if outputFile != "" {
		return writeLogsToFile(outputFile, logs)
	}

	// Display logs in the specified format
	if output.GetFormat(cmd).IsSerialized() {
		return output.SerializedOutput(cmd, logs.Data)
	}

	// Human-readable format
	list := output.NewList(cmd)
	for _, log := range logs.Data {
		logOut := &logEntryOut{
			Timestamp: log.Timestamp,
			Level:     log.Level,
			TaskId:    log.TaskId,
			Message:   log.Message,
		}
		list.Add(logOut)
	}

	if len(logs.Data) == 0 {
		output.Println(c.Config.EnableColor, "No more logs for the current query")
		return nil
	}

	output.Printf(c.Config.EnableColor, "Found %d log entries for connector %s:\n\n", len(logs.Data), connectorId)
	return list.Print()
}

func writeLogsToFile(outputFile string, logs *ccloudv2.LoggingSearchResponse) error {
	// Open file in append mode, create if it doesn't exist
	file, err := os.OpenFile(outputFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", outputFile, err)
	}
	defer file.Close()

	// Write each log entry individually
	for _, log := range logs.Data {
		logEntry := &logEntryOut{
			Timestamp: log.Timestamp,
			Level:     log.Level,
			TaskId:    log.TaskId,
			Message:   log.Message,
		}

		data, err := json.Marshal(logEntry)
		if err != nil {
			return fmt.Errorf("failed to marshal log entry to JSON: %w", err)
		}

		// Write the log entry as a single line (JSONL format)
		if _, err := file.Write(data); err != nil {
			return fmt.Errorf("failed to write log entry to file %s: %w", outputFile, err)
		}

		// Add newline after each log entry
		if _, err := file.WriteString("\n"); err != nil {
			return fmt.Errorf("failed to write newline to file %s: %w", outputFile, err)
		}
	}

	output.Printf(false, "Appended %d log entries to file: %s\n", len(logs.Data), outputFile)
	return nil
}

func validateTimeFormat(timeStr string) error {
	_, err := time.Parse(time.RFC3339, timeStr)
	return err
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
