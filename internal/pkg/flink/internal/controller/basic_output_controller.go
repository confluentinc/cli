package controller

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
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

	totalAvailableChars := c.calcTotalAvailableChars()
	rows := c.getRows(totalAvailableChars)
	rawTable := c.createTable(rows)
	rawTable.Render()
}

func (c *BasicOutputController) calcTotalAvailableChars() int {
	numColumns := len(c.resultFetcher.GetMaterializedStatementResults().GetHeaders())
	tableBorderNumChars := 4
	separatorCharsPerColumn := 3
	alreadyOccupiedSpace := tableBorderNumChars + (numColumns-1)*separatorCharsPerColumn
	return c.getWindowSize() - alreadyOccupiedSpace
}

func (c *BasicOutputController) getRows(totalAvailableChars int) [][]string {
	materializedStatementResults := c.resultFetcher.GetMaterializedStatementResults()
	columnWidths := materializedStatementResults.GetMaxWidthPerColumn()
	columnWidths = results.GetTruncatedColumnWidths(columnWidths, totalAvailableChars)

	// add actual row data
	rows := make([][]string, materializedStatementResults.Size())
	materializedStatementResults.ForEach(func(rowIdx int, row *types.StatementResultRow) {
		formattedRow := make([]string, len(row.Fields))
		for colIdx, field := range row.Fields {
			formattedRow[colIdx] = results.TruncateString(field.ToString(), columnWidths[colIdx])
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
	rawTable.AppendBulk(rows)
	return rawTable
}
