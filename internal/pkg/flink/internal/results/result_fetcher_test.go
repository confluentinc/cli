package results

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
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

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestInitSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: ""}
	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestToggleTableMode() {
	s.resultFetcher.ToggleTableMode()

	require.True(s.T(), s.resultFetcher.IsTableMode())
}

func (s *ResultFetcherTestSuite) TestToggleRefreshResults() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil).AnyTimes()
	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.SetAutoRefreshCallback(func() {
		s.resultFetcher.ToggleRefresh() // stop fetch
	})
	s.resultFetcher.ToggleRefresh() // start fetch

	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(1 * time.Second)
	}
	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterError() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsRefreshRunning())
	require.Equal(s.T(), types.Failed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterNoMorePageToken() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsRefreshRunning())
	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsFailedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(nil, &types.StatementError{})

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Failed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageDoesNotUpdateStateWhenAlreadyCompleted() {
	mockStatement := types.ProcessedStatement{PageToken: ""}
	s.resultFetcher.setStatement(mockStatement)
	s.resultFetcher.refreshState.setState(types.Completed)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{PageToken: "NOT_EMPTY"}, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
	require.Equal(s.T(), mockStatement, s.resultFetcher.GetStatement())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.refreshState.setState(types.Failed)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPagePreservesRunningState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.refreshState.setState(types.Running)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.fetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Running, s.resultFetcher.GetFetchState())
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

func (s *ResultFetcherTestSuite) TestCloseShouldSetFetchStateToPaused() {
	s.resultFetcher.refreshState.setState(types.Running)

	s.resultFetcher.Close()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestCloseShouldDeleteRunningStatements() {
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	s.resultFetcher.setStatement(statement)
	done := make(chan bool)
	s.mockStore.EXPECT().DeleteStatement(statement.StatementName).Do(
		func(statementName string) {
			done <- true
		})

	s.resultFetcher.Close()
	<-done

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestGetResults() {
	mockStatement := getStatementWithResultsExample()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), &s.resultFetcher.materializedStatementResults, s.resultFetcher.GetMaterializedStatementResults())
}

func (s *ResultFetcherTestSuite) TestReturnHeadersFromStatementResults() {
	mockStatement := getStatementWithResultsExample()

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), s.resultFetcher.materializedStatementResults.GetHeaders(), mockStatement.StatementResults.GetHeaders())
}

func (s *ResultFetcherTestSuite) TestReturnHeadersFromResultSchema() {
	mockStatement := getStatementWithResultsExample()
	mockStatement.StatementResults.Headers = nil
	columnDetails := generators.MockResultColumns(2, 1).Example()
	mockStatement.ResultSchema = flinkgatewayv1alpha1.SqlV1alpha1ResultSchema{Columns: &columnDetails}
	headers := make([]string, len(mockStatement.ResultSchema.GetColumns()))
	for idx, column := range mockStatement.ResultSchema.GetColumns() {
		headers[idx] = column.GetName()
	}

	s.resultFetcher.Init(mockStatement)

	require.Equal(s.T(), headers, s.resultFetcher.materializedStatementResults.GetHeaders())
}
