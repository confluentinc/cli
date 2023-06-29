package controller

import (
	"os"

	"github.com/olekukonko/tablewriter"

	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type OutputController struct {
	tableController types.TableControllerInterface
}

func NewOutputController(tableController types.TableControllerInterface) types.OutputControllerInterface {
	return &OutputController{
		tableController: tableController,
	}
}

func (c *OutputController) HandleStatementResults(processedStatement types.ProcessedStatement, windowSize int) {
	// decide if we want to display results using TView or just a plain table
	if shouldUseTView(processedStatement) {
		c.tableController.Init(processedStatement)
		c.tableController.Start()
		return
	}

	c.printResultToSTDOUT(processedStatement.StatementResults, windowSize)
}

func shouldUseTView(statement types.ProcessedStatement) bool {
	// only use view for non-local statements, that have more than one row and more than one column
	if statement.IsLocalStatement {
		return false
	}
	if statement.PageToken != "" {
		return true
	}
	return len(statement.StatementResults.GetHeaders()) > 1 && len(statement.StatementResults.GetRows()) > 1
}

func (c *OutputController) printResultToSTDOUT(statementResults *types.StatementResults, windowSize int) {
	if statementResults == nil || len(statementResults.Headers) == 0 || len(statementResults.Rows) == 0 {
		utils.OutputWarn("The server returned empty rows for this statement.")
		return
	}

	totalAvailableChars := c.calcTotalAvailableChars(len(statementResults.Headers), windowSize)
	rows := c.getRows(statementResults, totalAvailableChars)
	rawTable := c.createTable(statementResults.Headers, rows)
	rawTable.Render()
}

func (c *OutputController) calcTotalAvailableChars(numColumns int, windowSize int) int {
	tableBorderNumChars := 4
	separatorCharsPerColumn := 3
	alreadyOccupiedSpace := tableBorderNumChars + (numColumns-1)*separatorCharsPerColumn
	return windowSize - alreadyOccupiedSpace
}

func (c *OutputController) getRows(statementResults *types.StatementResults, totalAvailableChars int) [][]string {
	materializedStatementResults := types.NewMaterializedStatementResults(statementResults.GetHeaders(), results.MaxResultsCapacity)
	materializedStatementResults.Append(statementResults.GetRows()...)
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

func (c *OutputController) createTable(headers []string, rows [][]string) *tablewriter.Table {
	rawTable := tablewriter.NewWriter(os.Stdout)
	rawTable.SetAutoFormatHeaders(false)
	rawTable.SetHeader(headers)
	rawTable.AppendBulk(rows)
	return rawTable
}
