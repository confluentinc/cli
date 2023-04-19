package controller

import (
	"testing"

	"github.com/confluentinc/flink-sql-client/components"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/gdamore/tcell/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestAppInputCapture(t *testing.T) {
	mockAppController := NewMockApplicationControllerInterface(gomock.NewController(t))
	mockInputController := NewMockInputControllerInterface(gomock.NewController(t))
	mockStore := NewMockStoreInterface(gomock.NewController(t))
	table := components.CreateTable()
	tableController := NewTableController(table, mockStore, mockAppController)
	tableController.SetInputController(mockInputController)
	tc := tableController

	t.Run("Test Q", func(t *testing.T) {
		// Given
		input := tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)

		// When
		result := tc.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test Ctrl Q", func(t *testing.T) {
		// Given
		input := tcell.NewEventKey(tcell.KeyCtrlQ, rune(0), tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)

		// When
		result := tc.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test Escape", func(t *testing.T) {
		// Given
		input := tcell.NewEventKey(tcell.KeyEscape, rune(0), tcell.ModNone)
		mockAppController.EXPECT().SuspendOutputMode(gomock.Any()).Times(1)

		// When
		result := tc.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test N", func(t *testing.T) {
		// Given
		input := tcell.NewEventKey(tcell.KeyRune, 'N', tcell.ModNone)
		mockStatement := types.ProcessedStatement{
			StatementName: "test",
		}
		tviewApp := tview.NewApplication()
		mockStore.EXPECT().FetchStatementResults(gomock.Any()).Return(&mockStatement, nil).Times(1)
		mockAppController.EXPECT().TView().Return(tviewApp).Times(1)

		// When
		result := tc.AppInputCapture(input)

		// Then
		assert.Nil(t, result)
	})

	t.Run("Test Ctrl-C", func(t *testing.T) {
		input := tcell.NewEventKey(tcell.KeyCtrlC, rune(0), tcell.ModNone)
		result := tc.AppInputCapture(input)
		assert.Nil(t, result)
	})

	t.Run("Test default case", func(t *testing.T) {
		// Test a case when the event is neither 'Q', 'N', Ctrl-C, nor Escape
		// When we return the event, it's forwarded to tview
		input := tcell.NewEventKey(tcell.KeyF1, rune(0), tcell.ModNone)
		result := tc.AppInputCapture(input)
		assert.NotNil(t, result)
		assert.Equal(t, input, result)
	})
}
