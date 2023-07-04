package controller

import (
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

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
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat)
	s.resultFetcher.EXPECT().GetStatement().Return(types.ProcessedStatement{})
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.basicOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestRunInteractiveInputShouldNotPrintNoRowsWhenStatusDetailAvailable() {
	mat := types.NewMaterializedStatementResults([]string{}, 10)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat)
	s.resultFetcher.EXPECT().GetStatement().Return(types.ProcessedStatement{StatusDetail: "Created table 'test'"})
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.basicOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintTable() {
	executedStatementWithResults := getStatementWithResultsExample()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat).Times(4)

	stdout := test.RunAndCaptureSTDOUT(s.T(), s.basicOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func getStatementWithResultsExample() types.ProcessedStatement {
	statement := types.ProcessedStatement{
		StatementName: "example-statement",
		ResultSchema:  flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{},
		StatementResults: &types.StatementResults{
			Headers: []string{"Count"},
			Rows:    []types.StatementResultRow{},
		},
	}
	for i := 0; i < 10; i++ {
		row := types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(i),
				},
			},
		}
		statement.StatementResults.Rows = append(statement.StatementResults.Rows, row)
	}
	return statement
}
