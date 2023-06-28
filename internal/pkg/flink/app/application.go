package app

import (
	"os"

	"github.com/rivo/tview"
	"golang.org/x/term"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type Application struct {
	history             *history.History
	appController       types.ApplicationControllerInterface
	inputController     types.InputControllerInterface
	statementController types.StatementControllerInterface
	resultsController   types.OutputControllerInterface
	authenticated       func() error
}

func StartApp(client ccloudv2.GatewayClientInterface, authenticated func() error, appOptions types.ApplicationOptions) {
	// Load history of previous commands from cache file
	history := history.LoadHistory()

	// Create Components
	tviewApp := tview.NewApplication()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := controller.NewApplicationController(tviewApp, history)

	// Store used to process statements and store local properties
	store := store.NewStore(client, appController.ExitApplication, &appOptions)

	stdinBefore := getStdin()
	consoleParser := getConsoleParser()
	appController.AddCleanupFunction(func() {
		tearDownConsoleParser(consoleParser)
		restoreStdin(stdinBefore)
	})

	// Instantiate Component Controllers
	fetchController := controller.NewFetchController(store)
	tableController := controller.NewTableController(fetchController)
	inputController := controller.NewInputController(appController, history)
	statementController := controller.NewStatementController(appController, store, consoleParser)
	resultsController := controller.NewOutputController(tableController)

	app := Application{
		history:             history,
		appController:       appController,
		inputController:     inputController,
		statementController: statementController,
		resultsController:   resultsController,
		authenticated:       authenticated,
	}
	components.PrintWelcomeHeader()
	app.readEvalPrintLoop()
}

func getStdin() *term.State {
	state, err := term.GetState(int(os.Stdin.Fd()))
	if err != nil {
		log.CliLogger.Warnf("Couldn't get stdin state with term.GetState. Error: %v\n", err)
		return nil
	}
	return state
}

func getConsoleParser() prompt.ConsoleParser {
	consoleParser := prompt.NewStandardInputParser()
	err := consoleParser.Setup()
	if err != nil {
		log.CliLogger.Warnf("Couldn't setup console parser. Error: %v\n", err)
	}
	return consoleParser
}

func tearDownConsoleParser(consoleParser prompt.ConsoleParser) {
	err := consoleParser.TearDown()
	if err != nil {
		log.CliLogger.Warnf("Couldn't tear down console parser. Error: %v\n", err)
	}
}

func restoreStdin(state *term.State) {
	if state != nil {
		_ = term.Restore(int(os.Stdin.Fd()), state)
	}
}

func (a *Application) readEvalPrintLoop() {
	for a.isAuthenticated() {
		userInput := a.inputController.GetUserInput()
		if a.inputController.IsSpecialInput(userInput) {
			continue
		}
		a.history.Append(userInput)

		executedStatement, err := a.statementController.ExecuteStatement(userInput)
		if err != nil {
			continue
		}

		a.resultsController.HandleStatementResults(*executedStatement, a.inputController.GetWindowWidth())
	}
}

func (a *Application) isAuthenticated() bool {
	if authErr := a.authenticated(); authErr != nil {
		utils.OutputErrf("Error: %v\n", authErr)
		a.appController.ExitApplication()
		return false
	}
	return true
}
