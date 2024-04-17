package controller

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/confluentinc/cli/v3/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

type BasicOutputController struct {
	resultFetcher types.ResultFetcherInterface
	getWindowSize func() int
}

func NewBasicOutputController(resultFetcher types.ResultFetcherInterface, getWindowWidth func() int) types.OutputControllerInterface {
	return &BasicOutputController{
		resultFetcher: resultFetcher,
		getWindowSize: getWindowWidth,
	}
}

func (c *BasicOutputController) VisualizeResults() {
	c.printResultToSTDOUT()
}

func (c *BasicOutputController) printResultToSTDOUT() {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	rowsAreEmpty := len(materializedStatementResults.GetHeaders()) == 0 || materializedStatementResults.Size() == 0
	if rowsAreEmpty {
		if c.resultFetcher.GetStatement().StatusDetail == "" {
			utils.OutputWarn("The server returned empty rows for this statement.")
		}
		return
	}

	rows := c.getRows()
	rawTable := c.createTable(rows)
	rawTable.SetAlignment(tablewriter.ALIGN_LEFT)
	rawTable.Render()
}

func (c *BasicOutputController) getRows() [][]string {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()

	// add actual row data
	rows := make([][]string, materializedStatementResults.Size())
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		formattedRow := make([]string, len(row.Fields))
		for colIdx, field := range row.Fields {
			formattedRow[colIdx] = field.ToString()
		}
		rows[rowIdx] = formattedRow
	})
	return rows
}

func (c *BasicOutputController) createTable(rows [][]string) *tablewriter.Table {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetAutoFormatHeaders(false)
	rawTable.SetHeader(materializedStatementResults.GetHeaders())
	rawTable.SetAutoWrapText(true)
	rawTable.AppendBulk(rows)
	return rawTable
}
