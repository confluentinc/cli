package app

import (
	"sync"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/components"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/history"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/store"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

type Application struct {
	history                     *history.History
	store                       types.StoreInterface
	resultFetcher               types.ResultFetcherInterface
	appController               types.ApplicationControllerInterface
	inputController             types.InputControllerInterface
	statementController         types.StatementControllerInterface
	interactiveOutputController types.OutputControllerInterface
	basicOutputController       types.OutputControllerInterface
	tokenRefreshFunc            func() error
}

var mutex sync.Mutex

func synchronizedTokenRefresh(tokenRefreshFunc func() error) func() error {
	return func() error {
		mutex.Lock()
		defer mutex.Unlock()

		return tokenRefreshFunc()
	}
}

func StartApp(client ccloudv2.GatewayClientInterface, tokenRefreshFunc func() error, appOptions types.ApplicationOptions) {
	// Load history of previous commands from cache file
	historyStore := history.LoadHistory()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := controller.NewApplicationController(historyStore)

	// Store used to process statements and store local properties
	dataStore := store.NewStore(client, appController.ExitApplication, &appOptions, synchronizedTokenRefresh(tokenRefreshFunc))
	resultFetcher := results.NewResultFetcher(dataStore)

	stdinBefore := utils.GetStdin()
	consoleParser := utils.GetConsoleParser()
	appController.AddCleanupFunction(func() {
		utils.TearDownConsoleParser(consoleParser)
		utils.RestoreStdin(stdinBefore)
	})

	// Instantiate Component Controllers
	inputController := controller.NewInputController(historyStore)
	statementController := controller.NewStatementController(appController, dataStore, consoleParser)
	interactiveOutputController := controller.NewInteractiveOutputController(resultFetcher, appOptions.GetVerbose())
	basicOutputController := controller.NewBasicOutputController(resultFetcher, inputController.GetWindowWidth)

	app := Application{
		history:                     historyStore,
		store:                       dataStore,
		resultFetcher:               resultFetcher,
		appController:               appController,
		inputController:             inputController,
		statementController:         statementController,
		interactiveOutputController: interactiveOutputController,
		basicOutputController:       basicOutputController,
		tokenRefreshFunc:            synchronizedTokenRefresh(tokenRefreshFunc),
	}
	components.PrintWelcomeHeader()
	app.readEvalPrintLoop()
}

func (a *Application) readEvalPrintLoop() {
	for a.isAuthenticated() {
		userInput := a.inputController.GetUserInput()
		if a.inputController.HasUserEnabledReverseSearch() {
			a.inputController.StartReverseSearch()
			continue
		}
		if a.inputController.HasUserInitiatedExit(userInput) {
			a.appController.ExitApplication()
			return
		}
		a.history.Append(userInput)

		executedStatement, err := a.statementController.ExecuteStatement(userInput)
		if err != nil {
			continue
		}

		// fetch first result page here to init result fetcher and to decide which OutputController to use
		executedStatementWithResults, err := a.store.FetchStatementResults(*executedStatement)
		if err != nil {
			continue
		}
		a.resultFetcher.Init(*executedStatementWithResults)
		a.getOutputController(*executedStatementWithResults).VisualizeResults()
	}
}

func (a *Application) isAuthenticated() bool {
	if authErr := a.tokenRefreshFunc(); authErr != nil {
		utils.OutputErrf("Error: %v\n", authErr)
		a.appController.ExitApplication()
		return false
	}
	return true
}

func (a *Application) getOutputController(processedStatementWithResults types.ProcessedStatement) types.OutputControllerInterface {
	// only use view for non-local statements, that have more than one row and more than one column
	if processedStatementWithResults.IsLocalStatement {
		return a.basicOutputController
	}
	if processedStatementWithResults.PageToken != "" {
		return a.interactiveOutputController
	}
	if len(processedStatementWithResults.StatementResults.GetHeaders()) > 1 && len(processedStatementWithResults.StatementResults.GetRows()) > 1 {
		return a.interactiveOutputController
	}

	return a.basicOutputController
}
