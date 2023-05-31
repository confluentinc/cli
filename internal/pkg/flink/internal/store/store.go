package store

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/output"
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
	client          ccloudv2.GatewayClientInterface
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
	statementObj, err := s.client.CreateStatement(statement, s.Properties)
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

		updatedStatement, err := s.waitForPendingStatement(ctx, statement.StatementName, timeout(s.Properties))

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

	// Process remote statements that are now running or completed
	pageToken := statement.PageToken
	runningNoTokenRetries := 5
	for i := 0; i < runningNoTokenRetries; i++ {
		// TODO: we need to retry a few times on transient errors
		statementResultObj, err := s.client.GetStatementResults(statement.StatementName, pageToken)
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
		extractedToken, err := extractPageToken(statementMetadata.GetNext())
		if err != nil {
			return nil, &types.StatementError{Msg: "Error: " + err.Error()}
		}
		statement.PageToken = extractedToken
		if statement.Status == types.COMPLETED || statement.PageToken != "" || len(statementResults.GetData()) > 0 {
			// We try a few times to get non-empty token for RUNNING statements
			break
		}
		time.Sleep(time.Millisecond * 300)
	}
	return &statement, nil
}

func (s *Store) DeleteStatement(statementName string) bool {
	if !s.demoMode {
		err := s.client.DeleteStatement(statementName)
		if err != nil {
			log.CliLogger.Warnf("Failed to delete the statement: %v", err)
			return false
		}
	}
	return true
}

func (s *Store) waitForPendingStatement(ctx context.Context, statementName string, timeout time.Duration) (*types.ProcessedStatement, *types.StatementError) {
	retries := 0
	waitTime := calcWaitTime(retries)
	elapsedWaitTime := time.Millisecond * 300
	// Variable used to we inform the user every 5 seconds that we're still fetching for results (waiting for them to be ready)
	lastProgressUpdateTime := time.Second * 0
	var capturedErrors []string
	for {
		select {
		case <-ctx.Done():
			return nil, &types.StatementError{Msg: "Result retrieval aborted. Statement will be deleted.", HttpResponseCode: 499}
		default:
			statementObj, err := s.client.GetStatement(statementName)
			if err != nil {
				return nil, &types.StatementError{Msg: "Error: " + err.Error()}
			}

			phase := types.PHASE(statementObj.Status.GetPhase())
			if phase != types.PENDING {
				return types.NewProcessedStatement(statementObj), nil
			}
		}

		if len(capturedErrors) > 5 {
			break
		}

		lastProgressUpdateTime += waitTime
		elapsedWaitTime += waitTime
		time.Sleep(waitTime)

		if lastProgressUpdateTime.Seconds() > 5 {
			lastProgressUpdateTime = time.Second * 0
			output.Printf("Fetching results... (Timeout %d/%d) \n", int(elapsedWaitTime.Seconds()), int(timeout.Seconds()))
		}
		waitTime = calcWaitTime(retries)

		if elapsedWaitTime > timeout {
			break
		}
		retries++
	}

	var errorsMsg string
	if len(capturedErrors) > 0 {
		errorsMsg = fmt.Sprintf(" Captured retriable errors: %s", strings.Join(capturedErrors, "; "))
	}

	return nil, &types.StatementError{Msg: fmt.Sprintf("Error: Statement is still pending after %f seconds.%s \n\nIf you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.", timeout.Seconds(), errorsMsg)}
}

func extractPageToken(nextUrl string) (string, error) {
	if nextUrl == "" {
		return "", nil
	}
	myUrl, err := url.Parse(nextUrl)
	if err != nil {
		return "", err
	}
	params, err := url.ParseQuery(myUrl.RawQuery)
	if err != nil {
		return "", err
	}
	return params.Get("page_token"), nil
}

func NewStore(client ccloudv2.GatewayClientInterface, exitApplication func(), appOptions *types.ApplicationOptions) StoreInterface {
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
