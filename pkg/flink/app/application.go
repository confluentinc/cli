package app

import (
	"sync"
	"time"

	"github.com/sourcegraph/jsonrpc2"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/components"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/history"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/store"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/lsp"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type Application struct {
	history                     *history.History
	userProperties              types.UserPropertiesInterface
	store                       types.StoreInterface
	resultFetcher               types.ResultFetcherInterface
	appController               types.ApplicationControllerInterface
	inputController             types.InputControllerInterface
	statementController         types.StatementControllerInterface
	interactiveOutputController types.OutputControllerInterface
	baseOutputController        types.OutputControllerInterface
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
	userProperties := store.NewUserProperties(&appOptions)
	dataStore := store.NewStore(gatewayClient, appController.ExitApplication, userProperties, &appOptions, synchronizedTokenRefreshFunc)
	resultFetcher := results.NewResultFetcher(dataStore)

	// Instantiate LSP
	handlerCh := make(chan *jsonrpc2.Request) // This is the channel used for the messages received by the language to be passed through to the input controller
	lspClient, _, err := lsp.NewInitializedLspClient(getAuthToken, appOptions.GetLSPBaseUrl(), appOptions.GetOrganizationId(), appOptions.GetEnvironmentId(), handlerCh)
	if err != nil {
		log.CliLogger.Errorf("Failed to connect to the language service. Check your network."+
			" If you're using private networking, you might still be able to submit queries. If that's the case and you"+
			"want to uso language features like autocompletion, error highlighting and up to date syntax highlighting,"+
			" you or your system admin might need to setup DNS resolution for both \"flink\" AND \"flinkpls\""+
			" (e.g. flinkpls.us-east-2.aws.private.confluent.cloud). Contact support for assistance. Error: %s", err.Error())
	}

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
			AuthToken:           getAuthToken(),
			Catalog:             userProperties.Get(config.KeyCatalog),
			Database:            userProperties.Get(config.KeyDatabase),
			ComputePoolId:       appOptions.GetComputePoolId(),
			LspDocumentUri:      lspClient.CurrentDocumentUri(),
			StatementProperties: userProperties.GetMaskedNonLocalProperties(),
		}
	})

	inputController := controller.NewInputController(historyStore, lspCompleter, handlerCh, true)
	statementController := controller.NewStatementController(appController, dataStore, consoleParser)
	interactiveOutputController := controller.NewInteractiveOutputController(components.NewTableView(), resultFetcher, userProperties, appOptions.GetVerbose())
	baseOutputController := controller.NewBaseOutputController(resultFetcher, inputController.GetWindowWidth, userProperties)

	app := Application{
		history:                     historyStore,
		userProperties:              userProperties,
		store:                       dataStore,
		resultFetcher:               resultFetcher,
		appController:               appController,
		inputController:             inputController,
		statementController:         statementController,
		interactiveOutputController: interactiveOutputController,
		baseOutputController:        baseOutputController,
		refreshToken:                synchronizedTokenRefreshFunc,
		reportUsage:                 reportUsageFunc,
		appOptions:                  appOptions,
	}
	components.PrintWelcomeHeader(appOptions)
	return app.readEvalPrintLoop()
}

func StartAppOnPrem(flinkCmfClient *flink.CmfRestClient, tokenRefreshFunc func() error, appOptions types.ApplicationOptions) error {
	synchronizedTokenRefreshFunc := synchronizedTokenRefresh(tokenRefreshFunc)

	// Load history of previous commands from cache file
	historyStore := history.LoadHistoryOnPrem()

	// Instantiate Application Controller - this is the top level controller that will be passed down to all other controllers
	// and should be used for functions that are not specific to a component
	appController := controller.NewApplicationController(historyStore)

	// Store used to process statements and store local properties
	userProperties := store.NewUserProperties(&appOptions)
	dataStore := store.NewStoreOnPrem(flinkCmfClient, appController.ExitApplication, userProperties, &appOptions, synchronizedTokenRefreshFunc)
	resultFetcher := results.NewResultFetcher(dataStore)

	stdinBefore := utils.GetStdin()
	consoleParser, err := utils.GetConsoleParser()
	if err != nil {
		utils.OutputErr("Error: failed to initialize console parser")
		return errors.NewErrorWithSuggestions("failed to initialize console parser", "Restart your shell session or try another terminal.")
	}
	appController.AddCleanupFunction(func() {
		utils.TearDownConsoleParser(consoleParser)
		utils.RestoreStdin(stdinBefore)
	})

	// Instantiate Component Controllers
	inputController := controller.NewInputController(historyStore, nil, nil, false)
	statementController := controller.NewStatementController(appController, dataStore, consoleParser)
	interactiveOutputController := controller.NewInteractiveOutputController(components.NewTableView(), resultFetcher, userProperties, appOptions.GetVerbose())
	baseOutputController := controller.NewBaseOutputController(resultFetcher, inputController.GetWindowWidth, userProperties)

	app := Application{
		history:                     historyStore,
		userProperties:              userProperties,
		store:                       dataStore,
		resultFetcher:               resultFetcher,
		appController:               appController,
		inputController:             inputController,
		statementController:         statementController,
		interactiveOutputController: interactiveOutputController,
		baseOutputController:        baseOutputController,
		refreshToken:                synchronizedTokenRefreshFunc,
		reportUsage:                 func() { /* on-prem does not support usage reporting */ },
		appOptions:                  appOptions,
	}
	components.PrintWelcomeHeaderOnPrem(appOptions)
	return app.readEvalPrintLoop()
}

func (a *Application) readEvalPrintLoop() error {
	run := utils.NewPanicRecoveryWithLimit(3, 3*time.Second)
	for a.isAuthenticated() {
		err := run.WithCustomPanicRecovery(a.readEvalPrint, a.panicRecovery)
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
		// Append executedStatement text since sensitive information such as password can be masked in server side
		if executedStatement.Statement != "" {
			a.history.Append(executedStatement.Statement)
		}
	}

	if !executedStatement.IsDryRunStatement() {
		a.resultFetcher.Init(*executedStatement)
		a.getOutputController(*executedStatement).VisualizeResults()
	}
}

func (a *Application) panicRecovery() {
	log.CliLogger.Warn("Internal error occurred. Executing panic recovery.")
	a.statementController.CleanupStatement()
	a.interactiveOutputController = controller.NewInteractiveOutputController(components.NewTableView(), a.resultFetcher, a.userProperties, a.appOptions.GetVerbose())
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
		return a.baseOutputController
	}
	if processedStatementWithResults.PageToken != "" || processedStatementWithResults.IsSelectStatement() {
		return a.interactiveOutputController
	}

	return a.baseOutputController
}
