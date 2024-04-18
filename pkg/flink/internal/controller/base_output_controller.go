package controller

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

type BaseOutputController struct {
	resultFetcher types.ResultFetcherInterface
	getWindowSize func() int
	outputFormat  config.OutputFormat
}

func NewStandardOutputController(resultFetcher types.ResultFetcherInterface, getWindowWidth func() int) types.OutputControllerInterface {
	return &BaseOutputController{
		resultFetcher: resultFetcher,
		getWindowSize: getWindowWidth,
		outputFormat:  config.OutputFormatStandard,
	}
}

func (c *BaseOutputController) VisualizeResults() {
	c.printResultToSTDOUT()
}

func (c *BaseOutputController) printResultToSTDOUT() {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	rowsAreEmpty := len(materializedStatementResults.GetHeaders()) == 0 || materializedStatementResults.Size() == 0
	if rowsAreEmpty {
		if c.resultFetcher.GetStatement().StatusDetail == "" {
			utils.OutputWarn("The server returned empty rows for this statement.")
		}
		return
	}

	totalAvailableChars := c.calcTotalAvailableChars()
	rows := c.getRows(totalAvailableChars)
	rawTable := c.createTable(rows)
	rawTable.SetAlignment(tablewriter.ALIGN_LEFT)
	rawTable.Render()
}

func (c *BaseOutputController) calcTotalAvailableChars() int {
	numColumns := len(c.resultFetcher.GetMaterializedStatementResults().GetHeaders())
	tableBorderNumChars := 4
	separatorCharsPerColumn := 3
	alreadyOccupiedSpace := tableBorderNumChars + (numColumns-1)*separatorCharsPerColumn
	return c.getWindowSize() - alreadyOccupiedSpace
}

func (c *BaseOutputController) getRows(totalAvailableChars int) [][]string {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	columnWidths := materializedStatementResults.GetMaxWidthPerColumn()
	columnWidths = results.GetTruncatedColumnWidths(columnWidths, totalAvailableChars)

	// add actual row data
	rows := make([][]string, materializedStatementResults.Size())
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		formattedRow := make([]string, len(row.Fields))
		for colIdx, field := range row.Fields {
			if c.outputFormat == config.OutputFormatPlainText {
				formattedRow[colIdx] = field.ToString()
			} else {
				formattedRow[colIdx] = results.TruncateString(field.ToString(), columnWidths[colIdx])
			}
		}
		rows[rowIdx] = formattedRow
	})
	return rows
}

func (c *BaseOutputController) withBorder() bool {
	if c.outputFormat == config.OutputFormatPlainText {
		return false
	} else {
		return true
	}
}

func (c *BaseOutputController) withColumnSeparator() string {
	if c.outputFormat == config.OutputFormatPlainText {
		return ""
	} else {
		return "|"
	}
}

func (c *BaseOutputController) createTable(rows [][]string) *tablewriter.Table {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetAutoFormatHeaders(false)
	rawTable.SetHeader(materializedStatementResults.GetHeaders())
	rawTable.SetAutoWrapText(false)
	rawTable.AppendBulk(rows)
	rawTable.SetBorder(c.withBorder())
	rawTable.SetColumnSeparator(c.withColumnSeparator())
	return rawTable
}
