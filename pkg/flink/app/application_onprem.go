package app

import (
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/components"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/controller"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/history"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/store"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type ApplicationOnPrem struct {
	history                     *history.History
	userProperties              types.UserPropertiesInterface
	store                       types.StoreInterfaceOnPrem
	resultFetcher               types.ResultFetcherInterfaceOnPrem
	appController               types.ApplicationControllerInterface
	inputController             types.InputControllerInterface
	statementController         types.StatementControllerInterfaceOnPrem
	interactiveOutputController types.OutputControllerInterface
	baseOutputController        types.OutputControllerInterface
	refreshToken                func() error
	appOptions                  types.ApplicationOptions
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
	resultFetcher := results.NewResultFetcherOnPrem(dataStore)

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
	inputController := controller.NewInputController(historyStore, nil, nil)
	statementController := controller.NewStatementControllerOnPrem(appController, dataStore, consoleParser)
	interactiveOutputController := controller.NewInteractiveOutputControllerOnPrem(components.NewTableView(), resultFetcher, userProperties, appOptions.GetVerbose())
	baseOutputController := controller.NewBaseOutputControllerOnPrem(resultFetcher, inputController.GetWindowWidth, userProperties)

	app := ApplicationOnPrem{
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
		appOptions:                  appOptions,
	}
	components.PrintWelcomeHeaderOnPrem(appOptions)
	return app.readEvalPrintLoop()
}

func (a *ApplicationOnPrem) readEvalPrintLoop() error {
	run := utils.NewPanicRecoveryWithLimit(3, 3*time.Second)
	for a.isAuthenticated() {
		err := run.WithCustomPanicRecovery(a.readEvalPrint, a.panicRecovery)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *ApplicationOnPrem) readEvalPrint() {
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

	if !executedStatement.IsDryRunStatement() {
		a.resultFetcher.Init(*executedStatement)
		a.getOutputController(*executedStatement).VisualizeResults()
	}
}

func (a *ApplicationOnPrem) panicRecovery() {
	log.CliLogger.Warn("Internal error occurred. Executing panic recovery.")
	a.statementController.CleanupStatement()
	a.interactiveOutputController = controller.NewInteractiveOutputControllerOnPrem(components.NewTableView(), a.resultFetcher, a.userProperties, a.appOptions.GetVerbose())
}

func (a *ApplicationOnPrem) isAuthenticated() bool {
	if err := a.refreshToken(); err != nil {
		utils.OutputErrf("Error: %v", err)
		a.appController.ExitApplication()
		return false
	}
	return true
}

func (a *ApplicationOnPrem) getOutputController(processedStatementWithResults types.ProcessedStatement) types.OutputControllerInterface {
	if processedStatementWithResults.IsLocalStatement {
		return a.baseOutputController
	}
	if processedStatementWithResults.PageToken != "" || processedStatementWithResults.IsSelectStatement() {
		return a.interactiveOutputController
	}

	return a.baseOutputController
}
