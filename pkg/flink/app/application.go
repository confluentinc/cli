package app

import (
	"sync"
	"time"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/errors"
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

func StartApp(gatewayClient ccloudv2.GatewayClientInterface, tokenRefreshFunc func() error, appOptions types.ApplicationOptions, reportUsageFunc func()) error {
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
	consoleParser, err := utils.GetConsoleParser()
	if err != nil {
		utils.OutputErr("Error: failed to initialize console parser")
		return errors.NewErrorWithSuggestions("failed to initialize console parser", "Restart your shell session or try another terminal.")
	}
	appController.AddCleanupFunction(func() {
		utils.TearDownConsoleParser(consoleParser)
		utils.RestoreStdin(stdinBefore)
		if lspClient != nil {
			lspClient.ShutdownAndExit()
		}
	})

	// Instantiate Component Controllers
	lspCompleter := lsp.LspCompleter(lspClient, func() lsp.CliContext {
		return lsp.CliContext{
			AuthToken:     getAuthToken(),
			Catalog:       dataStore.GetCurrentCatalog(),
			Database:      dataStore.GetCurrentDatabase(),
			ComputePoolId: appOptions.GetComputePoolId(),
		}
	})

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
	return app.readEvalPrintLoop()
}

func (a *Application) readEvalPrintLoop() error {
	run := utils.NewPanicRecovererWithLimit(3, 3*time.Second)
	for a.isAuthenticated() {
		err := run.WithCustomPanicRecovery(a.readEvalPrint, a.panicRecovery)()
		if err != nil {
			return err
		}
	}
	return nil
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

	executedStatement, err := a.statementController.ExecuteStatement(userInput)
	if err != nil {
		return
	}
	if !executedStatement.IsSensitiveStatement {
		a.history.Append(userInput)
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
