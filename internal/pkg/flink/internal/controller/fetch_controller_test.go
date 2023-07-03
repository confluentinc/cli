package controller

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type FetchControllerTestSuite struct {
	suite.Suite
	fetchController *FetchController
	mockStore       *mock.MockStoreInterface
}

func TestFetchControllerTestSuite(t *testing.T) {
	suite.Run(t, new(FetchControllerTestSuite))
}

func (s *FetchControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockStore = mock.NewMockStoreInterface(ctrl)
	s.fetchController = NewFetchController(s.mockStore).(*FetchController)
}

func (s *FetchControllerTestSuite) TestToggleTableMode() {
	s.fetchController.ToggleTableMode()

	require.True(s.T(), s.fetchController.IsTableMode())
}

func (s *FetchControllerTestSuite) TestToggleRefreshResults() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil).AnyTimes()
	s.fetchController.Init(mockStatement)

	done := make(chan bool)
	// schedule pause
	go func() {
		time.Sleep(2 * time.Second)
		s.fetchController.ToggleAutoRefresh()
		// Then
		require.Equal(s.T(), types.Paused, s.fetchController.GetFetchState())
		done <- true
	}()
	<-done
}

func (s *FetchControllerTestSuite) TestResultFetchStopsAfterError() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

	s.fetchController.Init(mockStatement)
	// wait for auto refresh to complete
	for s.fetchController.IsAutoRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.fetchController.IsAutoRefreshRunning())
	require.Equal(s.T(), types.Failed, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestResultFetchStopsAfterNoMorePageToken() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.fetchController.Init(mockStatement)
	// wait for auto refresh to complete
	for s.fetchController.IsAutoRefreshRunning() {
		time.Sleep(1 * time.Second)
	}

	require.False(s.T(), s.fetchController.IsAutoRefreshRunning())
	require.Equal(s.T(), types.Completed, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestFetchNextPageSetsFailedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.fetchController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(nil, &types.StatementError{})

	s.fetchController.fetchNextPage()

	require.Equal(s.T(), types.Failed, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestFetchNextPageSetsCompletedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.fetchController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	s.fetchController.fetchNextPage()

	require.Equal(s.T(), types.Completed, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestFetchNextPageReturnsWhenAlreadyCompleted() {
	s.fetchController.setFetchState(types.Completed)

	s.fetchController.fetchNextPage()

	require.Equal(s.T(), types.Completed, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.fetchController.setFetchState(types.Failed)
	s.fetchController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.fetchController.fetchNextPage()

	require.Equal(s.T(), types.Paused, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestFetchNextPagePreservesRunningState() {
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	s.fetchController.setFetchState(types.Running)
	s.fetchController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	s.fetchController.fetchNextPage()

	require.Equal(s.T(), types.Running, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestCloseShouldSetFetchStateToPaused() {
	s.fetchController.setFetchState(types.Running)

	s.fetchController.Close()

	require.Equal(s.T(), types.Paused, s.fetchController.GetFetchState())
}

func (s *FetchControllerTestSuite) TestCloseShouldDeleteRunningStatements() {
	statement := types.ProcessedStatement{
		StatementName: "test-statement",
		Status:        types.RUNNING,
	}
	s.fetchController.setStatement(statement)
	done := make(chan bool)
	s.mockStore.EXPECT().DeleteStatement(statement.StatementName).Do(
		func(statementName string) {
			done <- true
		})

	s.fetchController.Close()
	<-done

	require.Equal(s.T(), types.Paused, s.fetchController.GetFetchState())
}
