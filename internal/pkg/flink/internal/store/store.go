package store

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/internal/results"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
	"github.com/confluentinc/cli/internal/pkg/log"
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
	appOptions      *types.ApplicationOptions
}

func (s *Store) ProcessLocalStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	switch statementType := parseStatementType(statement); statementType {
	case SetStatement:
		return s.processSetStatement(statement)
	case ResetStatement:
		return s.processResetStatement(statement)
	case UseStatement:
		return s.processUseStatement(statement)
	case ExitStatement:
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

	// Process remote statements
	statementObj, err := s.client.CreateStatement(
		statement,
		s.appOptions.GetComputePoolId(),
		s.appOptions.GetIdentityPoolId(),
		s.propsDefault(s.Properties),
		s.appOptions.GetEnvId(),
		s.appOptions.GetOrgResourceId(),
	)
	if err != nil {
		statusDetail := s.getStatusDetail(statementObj)
		return nil, &types.StatementError{
			Msg:            err.Error(),
			FailureMessage: statusDetail,
		}
	}
	return types.NewProcessedStatement(statementObj), nil
}

func (s *Store) WaitPendingStatement(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	statementStatus := statement.Status
	if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
		updatedStatement, err := s.waitForPendingStatement(ctx, statement.StatementName, timeout(s.Properties))
		if err != nil {
			return nil, err
		}

		// Check for failed or cancelled statements
		statementStatus = updatedStatement.Status
		if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
			return nil, &types.StatementError{
				Msg:            fmt.Sprintf("Can't fetch results. Statement phase is: %s", statementStatus),
				FailureMessage: updatedStatement.StatusDetail,
			}
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

	// Process remote statements that are now running or completed
	pageToken := statement.PageToken
	runningNoTokenRetries := 5
	for i := 0; i < runningNoTokenRetries; i++ {
		// TODO: we need to retry a few times on transient errors
		statementResultObj, err := s.client.GetStatementResults(s.appOptions.GetEnvId(), statement.StatementName, s.appOptions.GetOrgResourceId(), pageToken)
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}

		statementResults := statementResultObj.GetResults()
		convertedResults, err := results.ConvertToInternalResults(statementResults.GetData(), statement.ResultSchema)
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
		}
		statement.StatementResults = convertedResults

		statementMetadata := statementResultObj.GetMetadata()
		extractedToken, err := extractPageToken(statementMetadata.GetNext())
		if err != nil {
			return nil, &types.StatementError{Msg: err.Error()}
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
	if err := s.client.DeleteStatement(s.appOptions.GetEnvId(), statementName, s.appOptions.GetOrgResourceId()); err != nil {
		log.CliLogger.Warnf("Failed to delete the statement: %v", err)
		return false
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
			statementObj, err := s.client.GetStatement(s.appOptions.GetEnvId(), statementName, s.appOptions.GetOrgResourceId())
			statusDetail := s.getStatusDetail(statementObj)
			if err != nil {
				return nil, &types.StatementError{
					Msg:            err.Error(),
					FailureMessage: statusDetail}
			}

			phase := types.PHASE(statementObj.Status.GetPhase())
			if phase != types.PENDING {
				processedStatement := types.NewProcessedStatement(statementObj)
				processedStatement.StatusDetail = statusDetail
				return processedStatement, nil
			}

			// if status.detail is filled we encountered a retryable server response
			if statusDetail != "" {
				capturedErrors = append(capturedErrors, statusDetail)
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
		errorsMsg = fmt.Sprintf("Captured retriable errors: %s", strings.Join(capturedErrors, "; "))
	}

	return nil, &types.StatementError{
		Msg: fmt.Sprintf("Statement is still pending after %f seconds.\n If you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.",
			timeout.Seconds()),
		FailureMessage: errorsMsg,
	}
}

func (s *Store) getStatusDetail(statementObj flinkgatewayv1alpha1.SqlV1alpha1Statement) string {
	status := statementObj.GetStatus()
	phase := types.PHASE(status.GetPhase())
	if phase != types.FAILED && phase != types.FAILING {
		return status.GetDetail()
	}

	if status.GetDetail() != "" {
		return status.GetDetail()
	}

	// if the statement is in FAILED or FAILING phase and the status detail field is empty we show the latest exception instead
	exceptionsResponse, _ := s.client.GetExceptions(s.appOptions.GetEnvId(), statementObj.Spec.GetStatementName(), s.appOptions.GetOrgResourceId())
	exceptions := exceptionsResponse.GetData()
	if len(exceptions) < 1 {
		return ""
	}

	// most recent exception is on top of the returned list
	return exceptions[0].GetStacktrace()
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
	return &Store{
		Properties:      appOptions.GetDefaultProperties(),
		client:          client,
		exitApplication: exitApplication,
		appOptions:      appOptions,
	}
}

// Set properties default values if not set by the user
// We probably want to refactor the keys names and where they are stored. Maybe also the default values.
func (s *Store) propsDefault(propsWithoutDefault map[string]string) map[string]string {
	properties := make(map[string]string)
	for key, value := range propsWithoutDefault {
		properties[key] = value
	}

	if _, ok := properties[config.ConfigKeyCatalog]; !ok {
		properties[config.ConfigKeyCatalog] = s.appOptions.GetEnvId()
	}
	if _, ok := properties[config.ConfigKeyDatabase]; !ok {
		properties[config.ConfigKeyDatabase] = s.appOptions.GetKafkaClusterId()
	}
	if _, ok := properties[config.ConfigKeyOrgResourceId]; !ok {
		properties[config.ConfigKeyOrgResourceId] = s.appOptions.GetOrgResourceId()
	}
	if _, ok := properties[config.ConfigKeyExecutionRuntime]; !ok {
		properties[config.ConfigKeyExecutionRuntime] = "streaming"
	}
	if _, ok := properties[config.ConfigKeyLocalTimeZone]; !ok {
		properties[config.ConfigKeyLocalTimeZone] = getLocalTimezone()
	}

	// Here we delete locally used properties before sending it to the backend
	delete(properties, config.ConfigKeyResultsTimeout)

	return properties
}
