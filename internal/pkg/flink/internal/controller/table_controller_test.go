package controller

import (
	"github.com/confluentinc/flink-sql-client/internal/results"
	"github.com/stretchr/testify/suite"
	"pgregory.net/rapid"
	"strconv"
	"testing"

	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/rivo/tview"

	"github.com/confluentinc/flink-sql-client/components"
	"github.com/confluentinc/flink-sql-client/test/mock"
	"github.com/gdamore/tcell/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type TableControllerTestSuite struct {
	suite.Suite
	mockAppController   *mock.MockApplicationControllerInterface
	mockInputController *mock.MockInputControllerInterface
	mockStore           *mock.MockStoreInterface
}

func TestTableControllerTestSuite(t *testing.T) {
	suite.Run(t, new(TableControllerTestSuite))
}

func (s *TableControllerTestSuite) SetupSuite() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()
	s.mockAppController = mock.NewMockApplicationControllerInterface(ctrl)
	s.mockInputController = mock.NewMockInputControllerInterface(ctrl)
	s.mockStore = mock.NewMockStoreInterface(ctrl)
}

func (s *TableControllerTestSuite) TestQ() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
	s.mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
	s.mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

	// When
	result := tableController.AppInputCapture(input)

	// Then
	assert.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestCtrlQ() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)
	s.mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
	s.mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

	// When
	result := tableController.AppInputCapture(input)

	// Then
	assert.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestEscape() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)
	s.mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
	s.mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

	// When
	result := tableController.AppInputCapture(input)

	// Then
	assert.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestCtrlC() {
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyCtrlC, rune(0), tcell.ModNone)
	result := tableController.AppInputCapture(input)
	assert.Nil(s.T(), result)
}

func (s *TableControllerTestSuite) TestM() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).AnyTimes()
	s.mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).AnyTimes()

	// When
	tableController.Init(mockStatement)
	before := tableController.materializedStatementResults.IsTableMode()
	result := tableController.AppInputCapture(input)
	after := tableController.materializedStatementResults.IsTableMode()

	// Then
	assert.Nil(s.T(), result)
	assert.NotEqual(s.T(), after, before)
}

func (s *TableControllerTestSuite) TestR() {
	// Given
	table := components.CreateTable()
	mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).AnyTimes()
	s.mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).AnyTimes()

	// When
	tableController.Init(mockStatement)
	before := tableController.isAutoRefreshRunning()
	result := tableController.AppInputCapture(input)
	after := tableController.isAutoRefreshRunning()

	// Then
	assert.Nil(s.T(), result)
	assert.NotEqual(s.T(), after, before)
}

func (s *TableControllerTestSuite) TestDefault() {
	// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
	// When we return the event, it's forwarded to tview
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := tableController.AppInputCapture(input)
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), input, result)
}

func (s *TableControllerTestSuite) TestEnter() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp).Times(3)

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
	assert.Equal(s.T(), 10, tableController.selectedRowIdx) //header row is at 0 and last row at 10
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

func (s *TableControllerTestSuite) TestQInRowView() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
		isRowViewOpen: true,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().ShowTableView()
	s.mockAppController.EXPECT().TView().Return(tviewApp)

	// When
	result := tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone))

	// Then
	assert.Nil(s.T(), result)
	assert.False(s.T(), tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestCtrlQInRowView() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
		isRowViewOpen: true,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().ShowTableView()
	s.mockAppController.EXPECT().TView().Return(tviewApp)

	// When
	result := tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone))

	// Then
	assert.Nil(s.T(), result)
	assert.False(s.T(), tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestEscapeInRowView() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
		isRowViewOpen: true,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().ShowTableView()
	s.mockAppController.EXPECT().TView().Return(tviewApp)

	// When
	result := tableController.AppInputCapture(tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone))

	// Then
	assert.Nil(s.T(), result)
	assert.False(s.T(), tableController.isRowViewOpen)
}

func (s *TableControllerTestSuite) TestSelectRow() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)
	tviewApp := tview.NewApplication()
	s.mockAppController.EXPECT().TView().Return(tviewApp)

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
		s.mockAppController.EXPECT().TView().Return(tviewApp).Times(3)
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

func (s *TableControllerTestSuite) TestDefaultInRowView() {
	// Given
	table := components.CreateTable()
	tableController := TableController{
		table:         table,
		appController: s.mockAppController,
		store:         s.mockStore,
		isRowViewOpen: true,
	}
	tableController.SetRunInteractiveInputCallback(s.mockInputController.RunInteractiveInput)

	// When
	input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
	result := tableController.AppInputCapture(input)

	// Then
	assert.NotNil(s.T(), result)
	assert.Equal(s.T(), input, result)
	assert.True(s.T(), tableController.isRowViewOpen)
}
