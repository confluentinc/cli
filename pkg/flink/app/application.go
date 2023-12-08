package app

import (
	"sync"
	"time"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/flink/components"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/history"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/store"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v3/pkg/flink/lsp"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
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
	refreshToken                func() error
	reportUsage                 func()
	appOptions                  types.ApplicationOptions
}

var mutex sync.Mutex

func synchronizedTokenRefresh(tokenRefreshFunc func() error) func() error {
	return func() error {
		mutex.Lock()
		defer mutex.Unlock()

		return tokenRefreshFunc()
	}
}

func StartApp(gatewayClient ccloudv2.GatewayClientInterface, tokenRefreshFunc func() error, appOptions types.ApplicationOptions, reportUsageFunc func()) {
	synchronizedTokenRefreshFunc := synchronizedTokenRefresh(tokenRefreshFunc)
	getAuthToken := func() string {
		if authErr := synchronizedTokenRefreshFunc(); authErr != nil {
			log.CliLogger.Warnf("Failed to refresh token: %v", authErr)
		}
		return gatewayClient.GetAuthToken()
	}

	// Load history of previous commands from cache file
	historyStore := history.LoadHistory()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := controller.NewApplicationController(historyStore)

	// Store used to process statements and store local properties
	dataStore := store.NewStore(gatewayClient, appController.ExitApplication, &appOptions, synchronizedTokenRefreshFunc)
	resultFetcher := results.NewResultFetcher(dataStore)

	// Instantiate lsp
	lspClient := lsp.NewLSPClientWS(getAuthToken, appOptions.GetLSPBaseUrl(), appOptions.GetOrganizationId(), appOptions.GetEnvironmentId())

	stdinBefore := utils.GetStdin()
	consoleParser := utils.GetConsoleParser()
	if consoleParser == nil {
		utils.OutputErr("Error: failed to initialize console parser")
		return
	}
	appController.AddCleanupFunction(func() {
		utils.TearDownConsoleParser(consoleParser)
		utils.RestoreStdin(stdinBefore)
		if lspClient != nil {
			lspClient.ShutdownAndExit()
		}
	})

	// Instantiate Component Controllers
	var lspCompleter prompt.Completer
	if appOptions.LSPEnabled {
		lspCompleter = lsp.LSPCompleter(lspClient, func() lsp.CliContext {
			return lsp.CliContext{
				AuthToken:     getAuthToken(),
				Catalog:       dataStore.GetCurrentCatalog(),
				Database:      dataStore.GetCurrentDatabase(),
				ComputePoolId: appOptions.GetComputePoolId(),
			}
		})
	}

	inputController := controller.NewInputController(historyStore, lspCompleter)
	statementController := controller.NewStatementController(appController, dataStore, consoleParser)
	interactiveOutputController := controller.NewInteractiveOutputController(components.NewTableView(), resultFetcher, appOptions.GetVerbose())
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
		refreshToken:                synchronizedTokenRefreshFunc,
		reportUsage:                 reportUsageFunc,
		appOptions:                  appOptions,
	}
	components.PrintWelcomeHeader()
	app.readEvalPrintLoop()
}

func (a *Application) readEvalPrintLoop() {
	run := utils.NewPanicRecovererWithLimit(3, 3*time.Second)
	for a.isAuthenticated() {
		shouldExit := run.WithCustomPanicRecovery(a.readEvalPrint, a.panicRecovery)()
		if shouldExit {
			break
		}
	}
}

func (a *Application) readEvalPrint() {
	userInput := a.inputController.GetUserInput()
	if a.inputController.HasUserEnabledReverseSearch() {
		a.inputController.StartReverseSearch()
		return
	}
	if a.inputController.HasUserInitiatedExit(userInput) {
		a.appController.ExitApplication()
		return
	}
	a.history.Append(userInput)

	executedStatement, err := a.statementController.ExecuteStatement(userInput)
	if err != nil {
		return
	}

	a.resultFetcher.Init(*executedStatement)
	a.getOutputController(*executedStatement).VisualizeResults()
}

func (a *Application) panicRecovery() {
	log.CliLogger.Warn("Internal error occurred. Executing panic recovery.")
	a.statementController.CleanupStatement()
	a.interactiveOutputController = controller.NewInteractiveOutputController(components.NewTableView(), a.resultFetcher, a.appOptions.GetVerbose())
	a.reportUsage()
}

func (a *Application) isAuthenticated() bool {
	if err := a.refreshToken(); err != nil {
		utils.OutputErrf("Error: %v", err)
		a.appController.ExitApplication()
		return false
	}
	return true
}

func (a *Application) getOutputController(processedStatementWithResults types.ProcessedStatement) types.OutputControllerInterface {
	if processedStatementWithResults.IsLocalStatement {
		return a.basicOutputController
	}
	if processedStatementWithResults.PageToken != "" || processedStatementWithResults.IsSelectStatement {
		return a.interactiveOutputController
	}

	return a.basicOutputController
}
