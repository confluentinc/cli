package results

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"
	"pgregory.net/rapid"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
)

type ResultFetcherTestSuite struct {
	suite.Suite
	resultFetcher *ResultFetcher
	mockStore     *mock.MockStoreInterface
}

func TestResultFetcherTestSuite(t *testing.T) {
	suite.Run(t, new(ResultFetcherTestSuite))
}

func (s *ResultFetcherTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockStore = mock.NewMockStoreInterface(ctrl)
	s.resultFetcher = NewResultFetcher(s.mockStore).(*ResultFetcher)
}

func (s *ResultFetcherTestSuite) TestInitSetsPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestInitSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: ""}
	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestToggleTableMode() {
	s.resultFetcher.ToggleTableMode()

	require.True(s.T(), s.resultFetcher.IsTableMode())
}

func (s *ResultFetcherTestSuite) TestToggleRefreshResults() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil).AnyTimes()
	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.SetRefreshCallback(func() {
		s.resultFetcher.ToggleRefresh() // stop fetch
	})
	s.resultFetcher.ToggleRefresh() // start fetch

	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(time.Second)
	}
	require.Equal(s.T(), types.Paused, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterError() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsRefreshRunning())
	require.Equal(s.T(), types.Failed, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterNoMorePageToken() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsRefreshRunning())
	require.Equal(s.T(), types.Completed, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsFailedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(nil, &types.StatementError{})

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Failed, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageDoesNotUpdateStateWhenAlreadyCompleted() {
	mockStatement := types.ProcessedStatement{PageToken: ""}
	s.resultFetcher.setStatement(mockStatement)
	s.resultFetcher.refreshState.setState(types.Completed)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{PageToken: "NOT_EMPTY"}, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetRefreshState())
	require.Equal(s.T(), mockStatement, s.resultFetcher.GetStatement())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.refreshState.setState(types.Failed)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPagePreservesRunningState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.refreshState.setState(types.Running)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Running, s.resultFetcher.GetRefreshState())
}

func getStatementWithResultsExample() types.ProcessedStatement {
	return types.ProcessedStatement{
		StatementResults: &types.StatementResults{
			Headers: []string{"Test"},
			Rows: []types.StatementResultRow{{
				Operation: 0,
				Fields: []types.StatementResultField{
					types.AtomicStatementResultField{
						Type:  "INTEGER",
						Value: "1",
					},
				},
			}},
		},
		PageToken: "NOT_EMPTY",
	}
}

func (s *ResultFetcherTestSuite) TestCloseShouldSetRefreshStateToPaused() {
	s.resultFetcher.refreshState.setState(types.Running)

	s.resultFetcher.Close()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestCloseShouldDeleteRunningStatements() {
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	s.resultFetcher.setStatement(statement)
	done := make(chan bool)
	s.mockStore.EXPECT().StopStatement(statement.StatementName).Do(
		func(statementName string) {
			done <- true
		})

	s.resultFetcher.Close()
	<-done

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetRefreshState())
}

func (s *ResultFetcherTestSuite) TestGetResults() {
	mockStatement := getStatementWithResultsExample()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), &s.resultFetcher.materializedStatementResults, s.resultFetcher.GetMaterializedStatementResults())
}

func (s *ResultFetcherTestSuite) TestChangelogMode() {
	rapid.Check(s.T(), func(t *rapid.T) {
		// generate some results
		numColumns := rapid.IntRange(1, 10).Draw(t, maxNestingDepthLabel)
		results := generators.MockResults(numColumns, -1).Draw(t, "mock results")
		statementResults := results.StatementResults.Results.GetData()
		convertedResults, err := ConvertToInternalResults(statementResults, results.ResultSchema)
		require.NotNil(t, convertedResults)
		require.NoError(t, err)

		// test if in changelog mode all the rows are there and in the correct order
		materializedStatementResults := types.NewMaterializedStatementResults(convertedResults.GetHeaders(), 100, nil)
		materializedStatementResults.SetTableMode(false)
		materializedStatementResults.Append(convertedResults.GetRows()...)
		// in changelog mode we have an additional column "Operation"
		require.Equal(t, append([]string{"Operation"}, convertedResults.GetHeaders()...), materializedStatementResults.GetHeaders())
		require.Equal(t, len(convertedResults.GetRows()), materializedStatementResults.GetChangelogSize())
		iterator := materializedStatementResults.Iterator(false)
		for _, expectedRow := range convertedResults.GetRows() {
			actualRow := iterator.GetNext()
			operationField := types.AtomicStatementResultField{
				Type:  types.Varchar,
				Value: expectedRow.Operation.String(),
			}
			require.Equal(t, expectedRow.Operation, actualRow.Operation)
			// in changelog mode we have an additional column "Operation"
			require.Equal(t, append([]types.StatementResultField{operationField}, expectedRow.Fields...), actualRow.Fields)
		}
	})
}

func (s *ResultFetcherTestSuite) TestReturnHeadersFromStatementResults() {
	mockStatement := getStatementWithResultsExample()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), s.resultFetcher.materializedStatementResults.GetHeaders(), mockStatement.StatementResults.GetHeaders())
}

func (s *ResultFetcherTestSuite) TestReturnHeadersFromResultSchema_Cloud() {
	mockStatement := getStatementWithResultsExample()
	mockStatement.StatementResults.Headers = nil
	columnDetails := generators.MockResultColumns(2, 1).Example()
	mockStatement.Traits.FlinkGatewayV1StatementTraits = &flinkgatewayv1.SqlV1StatementTraits{Schema: &flinkgatewayv1.SqlV1ResultSchema{Columns: &columnDetails}}
	headers := mockStatement.Traits.GetColumnNames()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), headers, s.resultFetcher.materializedStatementResults.GetHeaders())
}

func (s *ResultFetcherTestSuite) TestReturnHeadersFromResultSchema_Onprem() {
	mockStatement := getStatementWithResultsExample()
	mockStatement.StatementResults.Headers = nil
	columnDetails := generators.MockResultColumnsOnPrem(2, 1).Example()
	mockStatement.Traits.CmfStatementTraits = &cmfsdk.StatementTraits{Schema: &cmfsdk.ResultSchema{Columns: columnDetails}}
	headers := mockStatement.Traits.GetColumnNames()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), headers, s.resultFetcher.materializedStatementResults.GetHeaders())
}
