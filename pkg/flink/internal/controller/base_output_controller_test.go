package controller

import (
	"strconv"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/store"
	"github.com/confluentinc/cli/v4/pkg/flink/test"
	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

type BasicOutputControllerTestSuite struct {
	suite.Suite
	standardOutputController  types.OutputControllerInterface
	plainTextOutputController types.OutputControllerInterface
	resultFetcher             *mock.MockResultFetcherInterface
}

func TestOutputControllerTestSuite(t *testing.T) {
	suite.Run(t, new(BasicOutputControllerTestSuite))
}

func (s *BasicOutputControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.resultFetcher = mock.NewMockResultFetcherInterface(ctrl)

	s.standardOutputController = NewBaseOutputController(s.resultFetcher, func() int {
		return 10
	}, userPropsWithStandardOutput())
	s.plainTextOutputController = NewBaseOutputController(s.resultFetcher, func() int {
		return 10
	}, userPropsWithPlainTextOutput())
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintNoRows() {
	mat := types.NewMaterializedStatementResults([]string{}, 10, nil)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat)
	s.resultFetcher.EXPECT().GetStatement().Return(types.ProcessedStatement{})
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.standardOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestRunInteractiveInputShouldNotPrintNoRowsWhenStatusDetailAvailable() {
	mat := types.NewMaterializedStatementResults([]string{}, 10, nil)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat)
	s.resultFetcher.EXPECT().GetStatement().Return(types.ProcessedStatement{StatusDetail: "Created table 'test'"})
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.standardOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintTable() {
	executedStatementWithResults := getStatementWithResultsExample()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10, nil)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat).Times(4)

	stdout := test.RunAndCaptureSTDOUT(s.T(), s.standardOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintNoRowsInPlainText() {
	mat := types.NewMaterializedStatementResults([]string{}, 10, nil)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat)
	s.resultFetcher.EXPECT().GetStatement().Return(types.ProcessedStatement{})
	stdout := test.RunAndCaptureSTDOUT(s.T(), s.plainTextOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func (s *BasicOutputControllerTestSuite) TestVisualizeResultsShouldPrintPlainTextTable() {
	executedStatementWithResults := getStatementWithResultsExample()
	mat := types.NewMaterializedStatementResults(executedStatementWithResults.StatementResults.GetHeaders(), 10, nil)
	mat.Append(executedStatementWithResults.StatementResults.GetRows()...)
	s.resultFetcher.EXPECT().GetMaterializedStatementResults().Return(&mat).Times(4)

	stdout := test.RunAndCaptureSTDOUT(s.T(), s.plainTextOutputController.VisualizeResults)

	cupaloy.SnapshotT(s.T(), stdout)
}

func getStatementWithResultsExample() types.ProcessedStatement {
	statement := types.ProcessedStatement{
		StatementName: "example-statement",
		Traits:        types.StatementTraits{},
		StatementResults: &types.StatementResults{
			Headers: []string{"Count"},
			Rows:    []types.StatementResultRow{},
		},
	}
	for i := 0; i < 10; i++ {
		row := types.StatementResultRow{
			Operation: types.Insert,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.Integer,
					Value: strconv.Itoa(i),
				},
			},
		}
		statement.StatementResults.Rows = append(statement.StatementResults.Rows, row)
	}
	return statement
}

func userPropsWithStandardOutput() types.UserPropertiesInterface {
	return store.NewUserPropertiesWithDefaults(map[string]string{config.KeyOutputFormat: string(config.OutputFormatStandard)}, map[string]string{})
}

func userPropsWithPlainTextOutput() types.UserPropertiesInterface {
	return store.NewUserPropertiesWithDefaults(map[string]string{config.KeyOutputFormat: string(config.OutputFormatPlainText)}, map[string]string{})
}
