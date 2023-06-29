package results

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

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

func (s *ResultFetcherTestSuite) TestToggleTableMode() {
	s.resultFetcher.ToggleTableMode()

	require.True(s.T(), s.resultFetcher.IsTableMode())
}

func (s *ResultFetcherTestSuite) TestToggleRefreshResults() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil).AnyTimes()
	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.SetAutoRefreshCallback(func() {
		s.resultFetcher.setFetchState(types.Paused)
	})
	s.resultFetcher.ToggleAutoRefresh()

	for s.resultFetcher.IsAutoRefreshRunning() {
		time.Sleep(1 * time.Second)
	}
	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterError() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleAutoRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsAutoRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsAutoRefreshRunning())
	require.Equal(s.T(), types.Failed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterNoMorePageToken() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.ToggleAutoRefresh()
	// wait for auto refresh to complete
	for s.resultFetcher.IsAutoRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.resultFetcher.IsAutoRefreshRunning())
	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsFailedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(nil, &types.StatementError{})

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Failed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageDoesNotUpdateStateWhenAlreadyCompleted() {
	mockStatement := types.ProcessedStatement{PageToken: ""}
	s.resultFetcher.setStatement(mockStatement)
	s.resultFetcher.setFetchState(types.Completed)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{PageToken: "NOT_EMPTY"}, nil)

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
	require.Equal(s.T(), mockStatement, s.resultFetcher.getStatement())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setFetchState(types.Failed)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPagePreservesRunningState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setFetchState(types.Running)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()

	require.Equal(s.T(), types.Running, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageOnUserInput() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)
	s.resultFetcher.Init(mockStatement)

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()
	// First nextPage returns statement with page token
	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())

	_, _ = s.resultFetcher.FetchNextPageAndUpdateState()
	// Second nextPage returns statement with empty page token, so state should be completed
	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestJumpToLiveResultsOnUserInput() {
	mockStatement := types.ProcessedStatement{
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
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{PageToken: "LAST"}, nil)

	// When
	s.resultFetcher.Init(mockStatement)

	// Then
	s.resultFetcher.JumpToLastPage()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
	require.Equal(s.T(), types.ProcessedStatement{PageToken: "LAST"}, s.resultFetcher.getStatement())
}

func (s *ResultFetcherTestSuite) TestCloseShouldSetFetchStateToPaused() {
	s.resultFetcher.setFetchState(types.Running)

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
