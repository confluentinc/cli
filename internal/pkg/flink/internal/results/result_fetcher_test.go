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

	done := make(chan bool)
	// schedule pause
	go func() {
		time.Sleep(2 * time.Second)
		s.resultFetcher.ToggleAutoRefresh()
		// Then
		require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
		done <- true
	}()
	<-done
}

func (s *ResultFetcherTestSuite) TestResultFetchStopsAfterError() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

	s.resultFetcher.Init(mockStatement)
	require.True(s.T(), s.resultFetcher.IsAutoRefreshRunning())
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
	require.True(s.T(), s.resultFetcher.IsAutoRefreshRunning())
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

	s.resultFetcher.FetchNextPage()

	require.Equal(s.T(), types.Failed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.resultFetcher.FetchNextPage()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageReturnsWhenAlreadyCompleted() {
	s.resultFetcher.setFetchState(types.Completed)

	s.resultFetcher.FetchNextPage()

	require.Equal(s.T(), types.Completed, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setFetchState(types.Failed)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.FetchNextPage()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPagePreservesRunningState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.resultFetcher.setFetchState(types.Running)
	s.resultFetcher.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.resultFetcher.FetchNextPage()

	require.Equal(s.T(), types.Running, s.resultFetcher.GetFetchState())
}

func (s *ResultFetcherTestSuite) TestFetchNextPageOnUserInput() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)
	s.resultFetcher.setFetchState(types.Completed) // need to manually set this so auto refresh doesn't start
	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.setFetchState(types.Paused)

	s.resultFetcher.FetchNextPage()
	// First nextPage returns statement with page token
	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())

	s.resultFetcher.FetchNextPage()
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
	s.resultFetcher.setFetchState(types.Completed) // need to manually set this so auto refresh doesn't start
	s.resultFetcher.Init(mockStatement)
	s.resultFetcher.setFetchState(types.Paused)

	// Then
	s.resultFetcher.JumpToLastPage()

	require.Equal(s.T(), types.Paused, s.resultFetcher.GetFetchState())
	require.Equal(s.T(), types.ProcessedStatement{PageToken: "LAST"}, s.resultFetcher.getStatement())
}
