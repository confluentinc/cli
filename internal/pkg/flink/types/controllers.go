package types

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type OutputMode string

var (
	GoPromptOutput OutputMode = "goprompt"
	TViewOutput    OutputMode = "tview"
)

type ApplicationControllerInterface interface {
	SuspendOutputMode(callback func())
	ToggleOutputMode()
	GetOutputMode() OutputMode
	ExitApplication()
	TView() *tview.Application
	ShowTableView()
	StartTView(layout tview.Primitive) error
	AddCleanupFunction(func()) ApplicationControllerInterface
}

type InputControllerInterface interface {
	GetUserInput() string
	IsSpecialInput(string) bool
	GetWindowWidth() int
}

type TableControllerInterface interface {
	AppInputCapture(event *tcell.EventKey) *tcell.EventKey
	Init(statement ProcessedStatement)
	SetRunInteractiveInputCallback(func())
}

type FetchControllerInterface interface {
	GetFetchState() FetchState
	IsTableMode() bool
	ToggleTableMode()
	ToggleAutoRefresh()
	IsAutoRefreshRunning() bool
	FetchNextPage()
	JumpToLastPage()
	GetHeaders() []string
	GetMaxWidthPerColumn() []int
	GetResultsIterator(bool) MaterializedStatementResultsIterator
	ForEach(func(rowIdx int, row *StatementResultRow))
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
}

type StatementControllerInterface interface {
	ExecuteStatement(statementToExecute string) (*ProcessedStatement, *StatementError)
}

type OutputControllerInterface interface {
	HandleStatementResults(processedStatement ProcessedStatement, windowSize int) bool
}
