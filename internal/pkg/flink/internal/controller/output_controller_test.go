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

type OutputControllerTestSuite struct {
	suite.Suite
	outputController types.OutputControllerInterface
	tableController  *mock.MockTableControllerInterface
}

func TestOutputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(OutputControllerTestSuite))
}

func (s *OutputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.tableController = mock.NewMockTableControllerInterface(ctrl)
	s.outputController = NewOutputController(s.tableController)
}

func (s *OutputControllerTestSuite) TestHandleStatementResultsShouldOpenTView() {
	processedStatement := types.ProcessedStatement{
		PageToken: "not-empty",
	}
	s.tableController.EXPECT().Init(processedStatement)

	isTViewUsed := s.outputController.HandleStatementResults(processedStatement, 10)

	require.True(s.T(), isTViewUsed)
}

func (s *OutputControllerTestSuite) TestHandleStatementResultsShouldPrintNoRows() {
	processedStatement := types.ProcessedStatement{}

	var isTViewUsed bool
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		isTViewUsed = s.outputController.HandleStatementResults(processedStatement, 10)
	})

	require.False(s.T(), isTViewUsed)
	require.Equal(s.T(), "The server returned empty rows for this statement.\n", stdout)
}

func (s *OutputControllerTestSuite) TestHandleStatementResultsShouldPrintTable() {
	processedStatement := getExampleStatementWithResults()

	var isTViewUsed bool
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		isTViewUsed = s.outputController.HandleStatementResults(processedStatement, 10)
	})

	require.False(s.T(), isTViewUsed)
	table := s.getStdoutTable(*processedStatement.StatementResults)
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

func (s *OutputControllerTestSuite) getStdoutTable(statementResults types.StatementResults) string {
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

func (s *OutputControllerTestSuite) TestShouldUseTView() {
	tests := []struct {
		name      string
		statement types.ProcessedStatement
		want      bool
	}{
		{
			name:      "local statement should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "local statement should not use TView even if unbounded",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: true},
			want:      false,
		},
		{
			name:      "non-local unbounded statement should always use TView",
			statement: types.ProcessedStatement{PageToken: "NOT_EMPTY", IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      true,
		},
		{
			name:      "statement with no results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false},
			want:      false,
		},
		{
			name:      "statement with empty results should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{}},
			want:      false,
		},
		{
			name: "statement with one column and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: false,
		},
		{
			name: "statement with two columns and one row should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows:    []types.StatementResultRow{{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}}},
			}},
			want: false,
		},
		{
			name: "statement with two columns and two rows should use TView",
			statement: types.ProcessedStatement{IsLocalStatement: false, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: true,
		},
		{
			name: "local statement with two columns and two rows should not use TView",
			statement: types.ProcessedStatement{IsLocalStatement: true, StatementResults: &types.StatementResults{
				Headers: []string{"Column 1", "Column 2"},
				Rows: []types.StatementResultRow{
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
					{Fields: []types.StatementResultField{types.AtomicStatementResultField{}}},
				},
			}},
			want: false,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shouldUseTView(tt.statement))
		})
	}
}
