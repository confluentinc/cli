package flink

import (
	"fmt"
	"net/url"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newStatementResultCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "result",
		Short: "Manage Flink SQL statement results.",
	}

	if cfg.IsCloudLogin() {
		pcmd.AddCloudFlag(cmd)
		pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)
		cmd.AddCommand(c.newStatementResultListCommand())
	}

	return cmd
}

type serializedResultOutput struct {
	Columns []string         `json:"columns" yaml:"columns"`
	Rows    []map[string]any `json:"rows" yaml:"rows"`
}

type statementResultData struct {
	Headers []string
	Rows    [][]string
}

func printStatementResults(cmd *cobra.Command, data *statementResultData) error {
	if data == nil || len(data.Rows) == 0 {
		if output.GetFormat(cmd).IsSerialized() {
			headers := []string{}
			if data != nil {
				headers = data.Headers
			}
			return output.SerializedOutput(cmd, &serializedResultOutput{
				Columns: headers,
				Rows:    []map[string]any{},
			})
		}
		fmt.Fprintln(cmd.OutOrStdout(), "No results found.")
		return nil
	}

	if output.GetFormat(cmd).IsSerialized() {
		rows := make([]map[string]any, len(data.Rows))
		for i, row := range data.Rows {
			rowMap := make(map[string]any)
			for j, val := range row {
				if j < len(data.Headers) {
					rowMap[data.Headers[j]] = val
				}
			}
			rows[i] = rowMap
		}
		return output.SerializedOutput(cmd, &serializedResultOutput{
			Columns: data.Headers,
			Rows:    rows,
		})
	}

	table := tablewriter.NewWriter(cmd.OutOrStdout())
	table.SetAutoFormatHeaders(false)
	table.SetHeader(data.Headers)
	table.SetAutoWrapText(false)
	table.SetBorder(false)

	for _, row := range data.Rows {
		table.Append(row)
	}

	table.Render()
	return nil
}

func fetchAllResults(client ccloudv2.GatewayClientInterface, envId, name, orgId string, schema flinkgatewayv1.SqlV1ResultSchema, maxRows int) (*statementResultData, error) {
	columns := schema.GetColumns()
	headers := make([]string, len(columns))
	for i, col := range columns {
		headers[i] = col.GetName()
	}

	var allRows [][]string
	pageToken := ""

	for {
		resp, err := client.GetStatementResults(envId, name, orgId, pageToken)
		if err != nil {
			return nil, err
		}

		rawData := resp.Results.GetData()
		for _, item := range rawData {
			resultItem, ok := item.(map[string]any)
			if !ok {
				continue
			}
			rowFields, _ := resultItem["row"].([]any)
			row := make([]string, len(headers))
			for j, field := range rowFields {
				if j < len(headers) {
					row[j] = fmt.Sprintf("%v", field)
				}
			}
			allRows = append(allRows, row)
		}

		if maxRows > 0 && len(allRows) >= maxRows {
			allRows = allRows[:maxRows]
			break
		}

		nextUrl := resp.Metadata.GetNext()
		nextToken, err := extractResultPageToken(nextUrl)
		if err != nil {
			return nil, err
		}
		if nextToken == "" {
			break
		}
		pageToken = nextToken
	}

	return &statementResultData{
		Headers: headers,
		Rows:    allRows,
	}, nil
}

func extractResultPageToken(nextUrl string) (string, error) {
	if nextUrl == "" {
		return "", nil
	}
	parsed, err := url.Parse(nextUrl)
	if err != nil {
		return "", err
	}
	params, err := url.ParseQuery(parsed.RawQuery)
	if err != nil {
		return "", err
	}
	return params.Get("page_token"), nil
}
