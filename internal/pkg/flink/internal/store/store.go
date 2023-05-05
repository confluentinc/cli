package store

import (
	"context"
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/confluentinc/flink-sql-client/internal/results"
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
	configKeyStatementOwner   = "statement-owner"
)

const MOCK_STATEMENTS_OUTPUT_DEMO = true

type StoreInterface interface {
	ProcessStatement(statement string) (*types.ProcessedStatement, *types.StatementError)
	FetchStatementResults(types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError)
	DeleteStatement(statementName string) bool
	WaitPendingStatement(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError)
}

type Store struct {
	Properties      map[string]string
	exitApplication func()
	client          GatewayClientInterface
	demoMode        bool
	mockCount       int
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
		s.exitApplication()
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
	if s.demoMode {

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

func (s *Store) WaitPendingStatement(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	// Process local statements

	statementStatus := statement.Status
	if statementStatus != types.COMPLETED && statementStatus != types.RUNNING && !s.demoMode {
		// Variable that controls how often we poll a pending statement
		const retries = 100
		const initialWaitTime = time.Millisecond * 300
		updatedStatement, err := s.waitForPendingStatement(ctx, statement.StatementName, retries, initialWaitTime)

		if err != nil {
			return nil, err
		}

		// Check for failed or cancelled statements
		statementStatus = updatedStatement.Status
		if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
			return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Can't fetch results. Statement phase is: %s", statementStatus)}
		}
		statement = *updatedStatement
	}

	return &statement, nil
}

func (s *Store) FetchStatementResults(statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	// Process local statements
	if statement.IsLocalStatement {
		return &statement, nil
	}

	statementStatus := statement.Status
	if statementStatus != types.COMPLETED && statementStatus != types.RUNNING && !s.demoMode {
		// Variable that controls how often we poll a pending statement
		const retries = 30
		const waitTime = time.Millisecond * 300
		updatedStatement, err := s.waitForPendingStatement(context.Background(), statement.StatementName, retries, waitTime)

		if err != nil {
			return nil, err
		}

		// Check for failed or cancelled statements
		statementStatus = updatedStatement.Status
		if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
			return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Can't fetch results. Statement phase is: %s", statementStatus)}
		}
		statement = *updatedStatement
	}
	// Process remote statements that are now running or completed
	statementResultObj, resp, err := s.client.GetStatementResults(context.Background(), statement.StatementName, statement.PageToken)
	err = processHttpErrors(resp, err)

	// TODO: Remove this once we have a real backend
	if s.demoMode {
		mockResults := generators.MockCount(s.mockCount)
		s.mockCount++
		statementResults := mockResults.StatementResults
		resultSchema := mockResults.ResultSchema
		convertedResults, err := results.ConvertToInternalResults(statementResults.Results.GetData(), resultSchema)
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}
		statement.StatementResults = convertedResults
		statement.PageToken = "TEST"
		return &statement, nil
	}
	if err != nil {
		return nil, &types.StatementError{Msg: err.Error()}
	}

	statementResults := statementResultObj.GetResults()
	convertedResults, err := results.ConvertToInternalResults(statementResults.GetData(), statement.ResultSchema)
	if err != nil {
		return nil, &types.StatementError{Msg: "Error: " + err.Error()}
	}
	statement.StatementResults = convertedResults

	statementMetadata := statementResultObj.GetMetadata()
	extractedToken, err := results.ExtractPageToken(statementMetadata.GetNext())
	if err != nil {
		return nil, &types.StatementError{Msg: "Error: " + err.Error()}
	}
	statement.PageToken = extractedToken
	return &statement, nil
}

func (s *Store) DeleteStatement(statementName string) bool {

	if !s.demoMode {
		httpResponse, err := s.client.DeleteStatement(context.Background(), statementName)

		if err != nil {
			log.Printf("Failed to delete the statement: %s", err.Error())
			return false
		}

		if httpResponse != nil && (httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300) {
			log.Printf("DeleteStatement returned unexpected status code: %v, with status: %s\n", httpResponse.StatusCode, httpResponse.Status)
			return false
		}
	}
	return true
}

func (s *Store) waitForPendingStatement(ctx context.Context, statementName string, retries int, waitTime time.Duration) (*types.ProcessedStatement, *types.StatementError) {
	var capturedErrors []string
	for i := 0; i < retries; i++ {
		select {
		case <-ctx.Done():
			return nil, &types.StatementError{Msg: "Result retrieval aborted. Statement will be deleted.", HttpResponseCode: 499}
		default:
			statementObj, httpResponse, err := s.client.GetStatement(context.Background(), statementName)

			if err != nil {
				return nil, &types.StatementError{Msg: "Error: " + err.Error()}
			}

			if httpResponse.StatusCode == http.StatusRequestTimeout ||
				httpResponse.StatusCode == http.StatusTooEarly ||
				httpResponse.StatusCode == http.StatusInternalServerError ||
				httpResponse.StatusCode == http.StatusServiceUnavailable ||
				httpResponse.StatusCode == http.StatusGatewayTimeout {
				capturedErrors = append(capturedErrors, statementObj.Status.GetDetail())
			} else if httpResponse.StatusCode < 200 || httpResponse.StatusCode >= 300 {
				return nil, &types.StatementError{Msg: "Error: " + statementObj.Status.GetDetail(), HttpResponseCode: httpResponse.StatusCode}
			} else {
				phase := types.PHASE(statementObj.Status.GetPhase())
				if phase != types.PENDING {
					return types.NewProcessedStatement(statementObj), nil
				}
			}
		}

		if len(capturedErrors) > 5 {
			break
		}

		time.Sleep(waitTime)

		// exponential backoff
		waitTime = (waitTime * 105) / 100
	}

	var errorsMsg string
	if len(capturedErrors) > 0 {
		errorsMsg = fmt.Sprintf(" Captured retriable errors: %s", strings.Join(capturedErrors, "; "))
	}
	return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Statement is still pending after %d retries.%s", retries, errorsMsg)}
}

func NewStore(client GatewayClientInterface, exitApplication func(), appOptions *types.ApplicationOptions) StoreInterface {
	defaultProperties := make(map[string]string)
	if appOptions != nil {
		if appOptions.DEFAULT_PROPERTIES != nil {
			defaultProperties = appOptions.DEFAULT_PROPERTIES
		}
	} else {
		appOptions = &types.ApplicationOptions{} // Initialize empty/default options
	}

	store := Store{
		Properties:      defaultProperties,
		client:          client,
		demoMode:        appOptions.MOCK_STATEMENTS_OUTPUT_DEMO,
		exitApplication: exitApplication,
	}

	return &store
}
