package controller

import (
	"os"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/olekukonko/tablewriter"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/flink/test"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type BasicOutputControllerTestSuite struct {
	suite.Suite
	basicOutputController types.OutputControllerInterface
	resultFetcher         *mock.MockResultFetcherInterface
}

func TestOutputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(BasicOutputControllerTestSuite))
}

func (s *BasicOutputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.resultFetcher = mock.NewMockResultFetcherInterface(ctrl)
	s.basicOutputController = NewBasicOutputController(s.resultFetcher, func() int {
		return 10
	})
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintNoRows() {
	mat := types.NewMaterializedStatementResults([]string{}, 10)
	s.resultFetcher.EXPECT().GetResults().Return(&mat).Times(1)
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.basicOutputController.VisualizeResults)

	require.Equal(s.T(), "The server returned empty rows for this statement.\n", stdout)
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintTable() {
	executedStatementWithResults := getExampleStatementWithResults()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)
	s.resultFetcher.EXPECT().GetResults().Return(&mat).Times(4)

	stdout := test.RunAndCaptureSTDOUT(s.T(), s.basicOutputController.VisualizeResults)

	table := s.getStdoutTable(*executedStatementWithResults.StatementResults)
	require.Equal(s.T(), table, stdout)
}

func getExampleStatementWithResults() types.ProcessedStatement {
	return types.ProcessedStatement{
		StatementResults: &types.StatementResults{
			Headers: []string{"column"},
			Rows: []types.StatementResultRow{
				{
					Operation: 0,
					Fields: []types.StatementResultField{
						types.AtomicStatementResultField{
							Type:  "INTEGER",
							Value: "0",
						},
					},
				},
			},
		},
	}
}

func (s *BasicOutputControllerTestSuite) getStdoutTable(statementResults types.StatementResults) string {
	return test.RunAndCaptureSTDOUT(s.T(), func() {
		rawTable := tablewriter.NewWriter(os.Stdout)
		rawTable.SetAutoFormatHeaders(false)
		rawTable.SetHeader(statementResults.Headers)
		for _, statementResultRow := range statementResults.GetRows() {
			row := make([]string, len(statementResultRow.Fields))
			for idx, field := range statementResultRow.Fields {
				row[idx] = field.ToString()
			}
			rawTable.Append(row)
		}
		rawTable.Render()
	})
}
