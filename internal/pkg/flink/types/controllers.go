package types

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/confluentinc/go-prompt"
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
	RunInteractiveInput()
	Prompt() prompt.IPrompt
	GetMaxCol() (int, error)
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
	GetHeaders() []string
	GetMaxWidthPerColumn() []int
	GetResultsIterator(bool) MaterializedStatementResultsIterator
	ForEach(func(rowIdx int, row *StatementResultRow))
	Init(statement ProcessedStatement)
	Close()
	SetAutoRefreshCallback(func())
	GetStatement() ProcessedStatement
	GetMaterializeStatementResults() *MaterializedStatementResults
}
