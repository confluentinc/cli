package controller

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type TableControllerTestSuite struct {
	suite.Suite
	mockAppController   *mock.MockApplicationControllerInterface
	mockInputController *mock.MockInputControllerInterface
	mockStore           *mock.MockStoreInterface
	dummyTViewApp       *tview.Application
}

func TestTableControllerTestSuite(t *testing.T) {
	suite.Run(t, new(TableControllerTestSuite))
}

func (s *TableControllerTestSuite) SetupTest() {
	ctrl := gomock.NewController(s.T())
	s.mockAppController = mock.NewMockApplicationControllerInterface(ctrl)
	s.mockInputController = mock.NewMockInputControllerInterface(ctrl)
	s.mockStore = mock.NewMockStoreInterface(ctrl)
	s.dummyTViewApp = tview.NewApplication()
}

func (s *TableControllerTestSuite) runWithRealTView(test func(*tview.Application)) {
	var once sync.Once
	tviewApp := tview.NewApplication()
	tviewApp.SetScreen(tcell.NewSimulationScreen(""))
	err := tviewApp.SetAfterDrawFunc(func(screen tcell.Screen) {
		if !screen.HasPendingEvent() {
			once.Do(func() {
				go func() {
					test(tviewApp)
					tviewApp.Stop()
				}()
			})
		}
	}).SetRoot(tview.NewBox(), true).Run()
	require.NoError(s.T(), err)
}

func (s *TableControllerTestSuite) TestCloseTableViewOnUserInput() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}

	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)},
		{name: "Test CtrlQ", input: tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)},
		{name: "Test Escape", input: tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			// Expected mock calls
			s.mockAppController.EXPECT().SuspendOutputMode(gomock.Any())

			// When
			result := tableController.AppInputCapture(testCase.input)

			// Then
			assert.Nil(s.T(), result)
		})
	}
}

func (s *TableControllerTestSuite) TestToggleTableModeOnUserInput() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
	s.mockAppController.EXPECT().TView().Return(tview.NewApplication()).Times(2)

	// When
	tableController.Init(mockStatement)
	before := tableController.materializedStatementResults.IsTableMode()
	result := tableController.AppInputCapture(input)
	after := tableController.materializedStatementResults.IsTableMode()

	// Then
	assert.Nil(s.T(), result)
	assert.NotEqual(s.T(), after, before)
}

func (s *TableControllerTestSuite) TestToggleRefreshResultsOnUserInput() {
	s.runWithRealTView(func(tview *tview.Application) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: s.mockAppController,
			store:         s.mockStore,
		}
		input := tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
		s.mockAppController.EXPECT().TView().Return(tview).AnyTimes()
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil).AnyTimes()

		// When
		tableController.Init(mockStatement)
		assert.True(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), running, tableController.getFetchState())

		done := make(chan bool)
		// schedule pause
		go func() {
			time.Sleep(2 * time.Second)
			result := tableController.AppInputCapture(input)
			// Then
			assert.Nil(s.T(), result)
			assert.False(s.T(), tableController.isAutoRefreshRunning())
			assert.Equal(s.T(), paused, tableController.getFetchState())
			done <- true
		}()
		<-done
	})
}

func (s *TableControllerTestSuite) TestNonSupportedUserInput() {
	// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
	// When we return the event, it's forwarded to tview
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := tableController.AppInputCapture(input)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), input, result)
}

func (s *TableControllerTestSuite) TestOpenRowViewOnUserInput() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	s.mockAppController.EXPECT().TView().Return(tview.NewApplication()).Times(3)

	headers := []string{"Count"}
	tableController.materializedStatementResults = results.NewMaterializedStatementResults(headers, 10)
	tableController.materializedStatementResults.SetTableMode(true)
	for i := 0; i < 10; i++ {
		tableController.materializedStatementResults.Append(types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(i),
				},
			},
		})
	}
	tableController.renderTable()

	// When
	// enter on the row
	result := tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

	// Then
	assert.Nil(s.T(), result)
	assert.True(s.T(), tableController.isRowViewOpen)
	assert.Equal(s.T(), 10, tableController.selectedRowIdx) // header row is at 0 and last row at 10
	// last row should be selected
	assert.Equal(s.T(), &types.StatementResultRow{
		Operation: types.INSERT,
		Fields: []types.StatementResultField{
			types.AtomicStatementResultField{
				Type:  types.INTEGER,
				Value: "9",
			},
		},
	}, tableController.materializedStatementResultsIterator.Value())
}

func (s *TableControllerTestSuite) TestCloseRowViewOnUserInput() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}

	testCases := []struct {
		name  string
		input *tcell.EventKey
	}{
		{name: "Test Q", input: tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)},
		{name: "Test CtrlQ", input: tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)},
		{name: "Test Escape", input: tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)},
	}

	for _, testCase := range testCases {
		s.T().Run(testCase.name, func(t *testing.T) {
			tableController.isRowViewOpen = true

			// Expected mock calls
			s.mockAppController.EXPECT().ShowTableView()
			s.mockAppController.EXPECT().TView().Return(tview.NewApplication())

			// When
			result := tableController.AppInputCapture(testCase.input)

			// Then
			assert.Nil(s.T(), result)
			assert.False(s.T(), tableController.isRowViewOpen)
		})
	}
}

func (s *TableControllerTestSuite) TestSelectRow() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	s.mockAppController.EXPECT().TView().Return(s.dummyTViewApp)

	headers := []string{"Count"}
	tableController.materializedStatementResults = results.NewMaterializedStatementResults(headers, 10)
	tableController.materializedStatementResults.SetTableMode(true)
	for i := 0; i < 10; i++ {
		tableController.materializedStatementResults.Append(types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(i),
				},
			},
		})
	}
	tableController.renderTable()

	rapid.Check(s.T(), func(t *rapid.T) {
		// Given
		rowToSelect := rapid.IntRange(-10, 20).Draw(t, "row to select")
		tableController.table.Select(rowToSelect, 0)
		// out of bounds handling
		if rowToSelect <= 0 {
			rowToSelect = 1
		}
		if rowToSelect >= tableController.table.GetRowCount() {
			rowToSelect = tableController.table.GetRowCount() - 1
		}
		s.mockAppController.EXPECT().TView().Return(s.dummyTViewApp).Times(3)
		s.mockAppController.EXPECT().ShowTableView()

		// When
		// enter on the row
		result := tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEnter, rune(0), tcell.ModNone))

		// Then
		assert.Nil(t, result)
		assert.True(t, tableController.isRowViewOpen)
		assert.Equal(t, rowToSelect, tableController.selectedRowIdx)
		// check if correct is selected
		assert.Equal(t, &types.StatementResultRow{
			Operation: types.INSERT,
			Fields: []types.StatementResultField{
				types.AtomicStatementResultField{
					Type:  types.INTEGER,
					Value: strconv.Itoa(rowToSelect - 1),
				},
			},
		}, tableController.materializedStatementResultsIterator.Value())

		// cleanup: exit row view
		result = tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone))
		assert.Nil(t, result)
	})
}

func (s *TableControllerTestSuite) TestNonSupportedUserInputInRowView() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
		isRowViewOpen: true,
	}

	// When
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := tableController.AppInputCapture(input)

	// Then
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), input, result)
	assert.True(s.T(), tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestResultFetchStopsAfterError() {
	s.runWithRealTView(func(tview *tview.Application) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: s.mockAppController,
			store:         s.mockStore,
		}
		s.mockAppController.EXPECT().TView().Return(tview).Times(2)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, &types.StatementError{Message: "error"})

		// When
		tableController.Init(mockStatement)
		assert.True(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), running, tableController.getFetchState())
		// wait for auto refresh to complete
		for tableController.isAutoRefreshRunning() {
			time.Sleep(1 * time.Second)
		}

		// Then
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), failed, tableController.getFetchState())
	})
}

func (s *TableControllerTestSuite) TestResultFetchStopsAfterNoMorePageToken() {
	s.runWithRealTView(func(tview *tview.Application) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: s.mockAppController,
			store:         s.mockStore,
		}
		s.mockAppController.EXPECT().TView().Return(tview).Times(2)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

		// When
		tableController.Init(mockStatement)
		assert.True(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), running, tableController.getFetchState())
		// wait for auto refresh to complete
		for tableController.isAutoRefreshRunning() {
			time.Sleep(1 * time.Second)
		}

		// Then
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), completed, tableController.getFetchState())
	})
}

func (s *TableControllerTestSuite) TestFetchNextPageSetsFailedState() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(nil, &types.StatementError{})

	// When
	assert.Equal(s.T(), paused, tableController.getFetchState())
	tableController.fetchNextPage()

	// Then
	assert.Equal(s.T(), failed, tableController.getFetchState())
}

func (s *TableControllerTestSuite) TestFetchNextPageSetsCompletedState() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

	// When
	assert.Equal(s.T(), paused, tableController.getFetchState())
	tableController.fetchNextPage()

	// Then
	assert.Equal(s.T(), completed, tableController.getFetchState())
}

func (s *TableControllerTestSuite) TestFetchNextPageReturnsWhenAlreadyCompleted() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.setFetchState(completed)

	// When
	assert.Equal(s.T(), completed, tableController.getFetchState())
	tableController.fetchNextPage()

	// Then
	assert.Equal(s.T(), completed, tableController.getFetchState())
}

func (s *TableControllerTestSuite) TestFetchNextPageChangesFailedToPausedState() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.setFetchState(failed)
	tableController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	// When
	assert.Equal(s.T(), failed, tableController.getFetchState())
	tableController.fetchNextPage()

	// Then
	assert.Equal(s.T(), paused, tableController.getFetchState())
}

func (s *TableControllerTestSuite) TestFetchNextPagePreservesRunningState() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.setFetchState(running)
	tableController.setStatement(mockStatement)
	s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)

	// When
	assert.Equal(s.T(), running, tableController.getFetchState())
	tableController.fetchNextPage()

	// Then
	assert.Equal(s.T(), running, tableController.getFetchState())
}

func (s *TableControllerTestSuite) TestFetchNextPageOnUserInput() {
	s.runWithRealTView(func(tview *tview.Application) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: s.mockAppController,
			store:         s.mockStore,
		}
		input := tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone)
		s.mockAppController.EXPECT().TView().Return(tview).Times(2)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{}, nil)

		// When
		tableController.setFetchState(completed) // need to manually set this so auto refresh doesn't start
		tableController.Init(mockStatement)
		tableController.setFetchState(paused)
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), paused, tableController.getFetchState())

		// Then
		// First N returns statement with page token
		assert.Nil(s.T(), tableController.AppInputCapture(input))
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), paused, tableController.getFetchState())

		// Second N returns statement with empty page token, so state should be completed
		assert.Nil(s.T(), tableController.AppInputCapture(input))
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), completed, tableController.getFetchState())
	})
}

func (s *TableControllerTestSuite) TestJumpToLiveResultsOnUserInput() {
	s.runWithRealTView(func(tview *tview.Application) {
		// Given
		table := components.CreateTable()
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
		tableController := TableController{
			table:         table,
			appController: s.mockAppController,
			store:         s.mockStore,
		}
		input := tcell.NewEventKey(tcell.KeyRune, 'L', tcell.ModNone)
		s.mockAppController.EXPECT().TView().Return(tview).Times(2)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&mockStatement, nil)
		s.mockStore.EXPECT().FetchStatementResults(mockStatement).Return(&types.ProcessedStatement{PageToken: "LAST"}, nil)

		// When
		tableController.setFetchState(completed) // need to manually set this so auto refresh doesn't start
		tableController.Init(mockStatement)
		tableController.setFetchState(paused)
		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), paused, tableController.getFetchState())

		// Then
		assert.Nil(s.T(), tableController.AppInputCapture(input))
		// wait for auto refresh to complete
		for tableController.getStatement().PageToken != "LAST" {
			time.Sleep(1 * time.Second)
		}

		assert.False(s.T(), tableController.isAutoRefreshRunning())
		assert.Equal(s.T(), paused, tableController.getFetchState())
		assert.Equal(s.T(), types.ProcessedStatement{PageToken: "LAST"}, tableController.getStatement())
	})
}
