package controller

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/flink-sql-client/pkg/converter"
	"github.com/confluentinc/flink-sql-client/pkg/types"
	"github.com/confluentinc/flink-sql-client/test/generators"
)

const (
	//ops
	configOpSet               = "SET"
	configOpUse               = "USE"
	configOpReset             = "RESET"
	configOpExit              = "EXIT"
	configOpUseCatalog        = "CATALOG"
	configStatementTerminator = ";"

	//keys
	configKeyCatalog          = "catalog"
	configKeyDatabase         = "default_database"
	configKeyOrgResourceId    = "org-resource-id"
	configKeyExecutionRuntime = "execution.runtime-mode"
)

const MOCK_STATEMENTS_OUTPUT_DEMO = true

type StoreInterface interface {
	ProcessStatement(statement string) (*types.ProcessedStatement, *types.StatementError)
	FetchStatementResults(types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError)
}

type Store struct {
	Properties          map[string]string
	ProcessedStatements []types.ProcessedStatement
	appController       ApplicationControllerInterface
	client              GatewayClientInterface
	appOptions          *ApplicationOptions
}

func (s *Store) ProcessLocalStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	switch statementType := parseStatementType(statement); statementType {
	case SET_STATEMENT:
		return s.processSetStatement(statement)
	case RESET_STATEMENT:
		return s.processResetStatement(statement)
	case USE_STATEMENT:
		return s.processUseStatement(statement)
	case EXIT_STATEMENT:
		s.appController.ExitApplication()
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *Store) ProcessStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	// We trim the statement here once so we don't have to do it in every function
	statement = strings.TrimSpace(statement)

	// Process local statements: set, use, reset
	result, sErr := s.ProcessLocalStatement(statement)
	if result != nil || sErr != nil {
		return result, sErr
	}

	// TODO: Remove this once we have a real backend
	if s.appOptions != nil && s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO {

		if !startsWithValidSQL(statement) {
			return nil, &types.StatementError{Msg: "Error: Invalid syntax '" + statement + "'. Please check your statement."}
		} else {
			return &types.ProcessedStatement{}, nil
		}
	}

	// Process remote statements
	statementObj, resp, err := s.client.CreateStatement(context.Background(), statement, s.Properties)
	err = processHttpErrors(resp, err)

	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}
	return types.NewProcessedStatement(statementObj), nil
}

func (s *Store) FetchStatementResults(statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	demoMode := s.appOptions != nil && s.appOptions.MOCK_STATEMENTS_OUTPUT_DEMO
	// Process local statements
	if statement.IsLocalStatement {
		return &statement, nil
	}

	statementStatus := statement.Status
	if statementStatus != types.COMPLETED && statementStatus != types.RUNNING && !demoMode {
		// Variable that controls how often we poll a pending statement
		const retries = 10
		const waitTime = time.Second * 1
		statement, err := s.waitForPendingStatement(statement.StatementName, retries, waitTime)

		// Check for errors
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}

		// Check for failed or cancelled statements
		statementStatus = statement.Status
		if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
			return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Can't fetch results. Statement phase is: %s", statementStatus)}
		}
	}
	// Process remote statements that are now running or completed
	statementResultObj, resp, err := s.client.GetStatementResults(context.Background(), statement.StatementName, statement.PageToken)
	err = processHttpErrors(resp, err)

	// TODO: Remove this once we have a real backend
	if demoMode {
		mockResults := generators.MockResults(5, 2).Example()
		statementResults := mockResults.StatementResults
		resultSchema := mockResults.ResultSchema
		convertedResults, err := converter.ConvertToInternalResults(statementResults.Results.GetData(), resultSchema)
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}
		statement.StatementResults = convertedResults
		return &statement, nil
	}
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}

	results := statementResultObj.GetResults()
	convertedResults, err := converter.ConvertToInternalResults(results.GetData(), statement.ResultSchema)
	if err != nil {
		return nil, &types.StatementError{Msg: "Error: " + err.Error()}
	}

	statement.StatementResults = convertedResults
	metadata := statementResultObj.GetMetadata()
	statement.PageToken = metadata.GetNext()
	return &statement, nil
}

func (s *Store) waitForPendingStatement(statementName string, retries int, waitTime time.Duration) (*types.ProcessedStatement, error) {

	for i := 0; i < retries; i++ {
		statementObj, httpResponse, err := s.client.GetStatement(context.Background(), statementName)

		if err != nil {
			return nil, err
		}

		if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
			return nil, &types.StatementError{Msg: "Error: " + statementObj.Status.GetDetail()}
		}

		phase := types.PHASE(statementObj.Status.GetPhase())
		if phase != types.PENDING {
			return types.NewProcessedStatement(statementObj), nil
		}

		time.Sleep(waitTime)
	}

	return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Statement is still pending after %d retries", retries)}
}

func NewStore(client GatewayClientInterface, appOptions *ApplicationOptions, appController ApplicationControllerInterface) StoreInterface {
	defaultProperties := make(map[string]string)

	if appOptions != nil && appOptions.DEFAULT_PROPERTIES != nil {
		defaultProperties = appOptions.DEFAULT_PROPERTIES
	}

	store := Store{
		Properties:    defaultProperties,
		client:        client,
		appOptions:    appOptions,
		appController: appController,
	}

	return &store
}
