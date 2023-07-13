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

type Store struct {
	Properties       UserProperties
	exitApplication  func()
	client           ccloudv2.GatewayClientInterface
	appOptions       *types.ApplicationOptions
	tokenRefreshFunc func() error
}

func (s *Store) authenticatedGatewayClient() ccloudv2.GatewayClientInterface {
	if authErr := s.tokenRefreshFunc(); authErr != nil {
		log.CliLogger.Warnf("Failed to refresh token: %v", authErr)
	}
	return s.client
}

func (s *Store) ProcessLocalStatement(statement string) (*types.ProcessedStatement, *types.StatementError) {
	defer s.persistUserProperties()
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
	statementObj, err := s.authenticatedGatewayClient().CreateStatement(
		statement,
		s.appOptions.GetComputePoolId(),
		s.appOptions.GetIdentityPoolId(),
		s.Properties.GetProperties(),
		s.appOptions.GetEnvironmentId(),
		s.appOptions.GetOrgResourceId(),
	)
	if err != nil {
		statusDetail := s.getStatusDetail(statementObj)
		return nil, &types.StatementError{
			Message:        err.Error(),
			FailureMessage: statusDetail,
		}
	}
	return types.NewProcessedStatement(statementObj), nil
}

func (s *Store) persistUserProperties() {
	if s.appOptions.GetContext() != nil {
		if err := s.appOptions.Context.SetCurrentFlinkCatalog(s.Properties.Get(config.ConfigKeyCatalog)); err != nil {
			log.CliLogger.Errorf("error setting current flink catalog: %v", err)
		}

		if err := s.appOptions.Context.SetCurrentFlinkDatabase(s.Properties.Get(config.ConfigKeyDatabase)); err != nil {
			log.CliLogger.Errorf("error setting current flink database: %v", err)
		}

		if err := s.appOptions.Context.Save(); err != nil {
			log.CliLogger.Errorf("error persisting user properties: %v", err)
		}
	}
}

func (s *Store) WaitPendingStatement(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	statementStatus := statement.Status
	if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
		updatedStatement, err := s.waitForPendingStatement(ctx, statement.StatementName, s.getTimeout())
		if err != nil {
			return nil, err
		}

		// Check for failed or cancelled statements
		statementStatus = updatedStatement.Status
		if statementStatus != types.COMPLETED && statementStatus != types.RUNNING {
			return nil, &types.StatementError{
				Message:        fmt.Sprintf("can't fetch results. Statement phase is: %s", statementStatus),
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
	statementResultObj, err := s.authenticatedGatewayClient().GetStatementResults(s.appOptions.GetEnvironmentId(), statement.StatementName, s.appOptions.GetOrgResourceId(), statement.PageToken)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}

	statementResults := statementResultObj.GetResults()
	convertedResults, err := results.ConvertToInternalResults(statementResults.GetData(), statement.ResultSchema)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	statement.StatementResults = convertedResults

	statementMetadata := statementResultObj.GetMetadata()
	extractedToken, err := extractPageToken(statementMetadata.GetNext())
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}
	statement.PageToken = extractedToken
	return &statement, nil
}

func (s *Store) DeleteStatement(statementName string) bool {
	if err := s.authenticatedGatewayClient().DeleteStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrgResourceId()); err != nil {
		log.CliLogger.Warnf("Failed to delete the statement: %v", err)
		return false
	}
	log.CliLogger.Infof("Successfully deleted statement: %s", statementName)
	return true
}

func (s *Store) waitForPendingStatement(ctx context.Context, statementName string, timeout time.Duration) (*types.ProcessedStatement, *types.StatementError) {
	retries := 0
	waitTime := calcWaitTime(retries)
	elapsedWaitTime := time.Millisecond * 300
	// Variable used to we inform the user every 5 seconds that we're still fetching for results (waiting for them to be ready)
	lastProgressUpdateTime := time.Second * 0
	var capturedErrors []string
	var phase types.PHASE
	capturedErrorsLimit := 5
	var getRequestDuration time.Duration
	for {
		select {
		case <-ctx.Done():
			return nil, &types.StatementError{Message: "result retrieval aborted. Statement will be deleted", HttpResponseCode: 499}
		default:
			start := time.Now()
			statementObj, err := s.authenticatedGatewayClient().GetStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrgResourceId())
			getRequestDuration = time.Since(start)

			statusDetail := s.getStatusDetail(statementObj)
			if err != nil {
				return nil, &types.StatementError{
					Message:        err.Error(),
					FailureMessage: statusDetail}
			}

			phase = types.PHASE(statementObj.Status.GetPhase())
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

		if len(capturedErrors) > capturedErrorsLimit {
			return nil, &types.StatementError{
				Message: fmt.Sprintf("the server can't process this statement right now, exiting after %d retries",
					len(capturedErrors)),
				FailureMessage: fmt.Sprintf("captured retryable errors: %s", strings.Join(capturedErrors, "; ")),
			}
		}

		if getRequestDuration > waitTime {
			lastProgressUpdateTime += getRequestDuration
			elapsedWaitTime += getRequestDuration
		} else {
			lastProgressUpdateTime += waitTime
			elapsedWaitTime += waitTime
			waitTime -= getRequestDuration
			time.Sleep(waitTime)
		}

		if int(lastProgressUpdateTime.Seconds()) > capturedErrorsLimit {
			lastProgressUpdateTime = time.Second * 0
			output.Printf("Waiting for statement to be ready. Statement phase is %s. (Timeout %ds/%ds) \n", phase, int(elapsedWaitTime.Seconds()), int(timeout.Seconds()))
		}
		waitTime = calcWaitTime(retries)

		if elapsedWaitTime > timeout {
			break
		}
		retries++
	}

	var errorsMsg string
	if len(capturedErrors) > 0 {
		errorsMsg = fmt.Sprintf("captured retryable errors: %s", strings.Join(capturedErrors, "; "))
	}

	return nil, &types.StatementError{
		Message: fmt.Sprintf("statement is still pending after %f seconds. If you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.",
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
	exceptionsResponse, err := s.authenticatedGatewayClient().GetExceptions(s.appOptions.GetEnvironmentId(), statementObj.Spec.GetStatementName(), s.appOptions.GetOrgResourceId())
	if err != nil {
		return ""
	}

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

func NewStore(client ccloudv2.GatewayClientInterface, exitApplication func(), appOptions *types.ApplicationOptions, tokenRefreshFunc func() error) types.StoreInterface {
	return &Store{
		Properties:       NewUserProperties(getDefaultProperties(appOptions)),
		client:           client,
		exitApplication:  exitApplication,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}
}

func getDefaultProperties(appOptions *types.ApplicationOptions) map[string]string {
	properties := map[string]string{
		config.ConfigKeyCatalog:       appOptions.GetEnvironmentName(),
		config.ConfigKeyDatabase:      appOptions.GetDatabase(),
		config.ConfigKeyLocalTimeZone: getLocalTimezone(),
	}

	return properties
}
