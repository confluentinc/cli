package controller

import (
	"testing"

	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/rivo/tview"

	"github.com/confluentinc/flink-sql-client/components"
	"github.com/gdamore/tcell/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAppInputCapture(t *testing.T) {
	mockAppController := NewMockApplicationControllerInterface(gomock.NewController(t))
	mockInputController := NewMockInputControllerInterface(gomock.NewController(t))
	mockStore := NewMockStoreInterface(gomock.NewController(t))

	t.Run("Test Q", func(t *testing.T) {
		// Given
		table := components.CreateTable()
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
		tviewApp := tview.NewApplication()
		mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
		mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

		// When
		result := tableController.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test Ctrl Q", func(t *testing.T) {
		// Given
		table := components.CreateTable()
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
		tviewApp := tview.NewApplication()
		mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
		mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

		// When
		result := tableController.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test Escape", func(t *testing.T) {
		// Given
		table := components.CreateTable()
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)
		tviewApp := tview.NewApplication()
		mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
		mockStore.EXPECT().DeleteStatement(gomock.Any()).Times(1)

		// When
		result := tableController.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	//t.Run("Test N", func(t *testing.T) {
	//	// Given
	//	input := tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone)
	//	mockStatement := types.ProcessedStatement{
	//		StatementName: "test",
	//	}
	//	tviewApp := tview.NewApplication()
	//	mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).Times(1)
	//	mockAppController.EXPECT().TView().Return(tviewApp).Times(1)
	//
	//	// When
	//	result := tableController.AppInputCapture(input)
	//
	//	// Then
	//	assert.Nil(t, result)
	//})

	t.Run("Test Ctrl-C", func(t *testing.T) {
		table := components.CreateTable()
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyCtrlC, rune(0), tcell.ModNone)
		result := tableController.AppInputCapture(input)
		assert.Nil(t, result)
	})

	t.Run("Test M", func(t *testing.T) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyRune, 'M', tcell.ModNone)
		tviewApp := tview.NewApplication()
		mockAppController.EXPECT().TView().Return(tviewApp).AnyTimes()
		mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).AnyTimes()

		// When
		tableController.Init(mockStatement)
		before := tableController.materializedStatementResults.IsTableMode()
		result := tableController.AppInputCapture(input)
		after := tableController.materializedStatementResults.IsTableMode()

		// Then
		assert.Nil(t, result)
		assert.NotEqual(t, after, before)
	})

	t.Run("Test R", func(t *testing.T) {
		// Given
		table := components.CreateTable()
		mockStatement := types.ProcessedStatement{PageToken: "NOT_EMPTY"}
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone)
		tviewApp := tview.NewApplication()
		mockAppController.EXPECT().TView().Return(tviewApp).AnyTimes()
		mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).AnyTimes()

		// When
		tableController.Init(mockStatement)
		before := tableController.isAutoRefreshRunning()
		result := tableController.AppInputCapture(input)
		after := tableController.isAutoRefreshRunning()

		// Then
		assert.Nil(t, result)
		assert.NotEqual(t, after, before)
	})

	t.Run("Test default case", func(t *testing.T) {
		// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
		// When we return the event, it's forwarded to tview
		table := components.CreateTable()
		tableController := TableController{
			table:         table,
			appController: mockAppController,
			store:         mockStore,
		}
		tableController.SetInputController(mockInputController)
		input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
		result := tableController.AppInputCapture(input)
		assert.NotNil(t, result)
		assert.Equal(t, input, result)
	})
}
