package store

import (
	"context"
	_ "embed"
	"fmt"
	"net/url"
	"strings"
	"time"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type Store struct {
	Properties       UserProperties
	exitApplication  func()
	client           ccloudv2.GatewayClientInterface
	appOptions       *types.ApplicationOptions
	tokenRefreshFunc func() error
}

func (s *Store) GetCurrentCatalog() string {
	return s.Properties.Get(config.KeyCatalog)
}

func (s *Store) GetCurrentDatabase() string {
	return s.Properties.Get(config.KeyDatabase)
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
	case QuitStatement:
		fallthrough
	case ExitStatement:
		s.exitApplication()
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *Store) persistUserProperties() {
	if s.appOptions.GetContext() != nil {
		if err := s.appOptions.Context.SetCurrentFlinkCatalog(s.Properties.Get(config.KeyCatalog)); err != nil {
			log.CliLogger.Errorf("error persisting current flink catalog: %v", err)
		}

		if err := s.appOptions.Context.SetCurrentFlinkDatabase(s.Properties.Get(config.KeyDatabase)); err != nil {
			log.CliLogger.Errorf("error persisting current flink database: %v", err)
		}

		if err := s.appOptions.Context.Save(); err != nil {
			log.CliLogger.Errorf("error persisting user properties: %v", err)
		}
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

	statementName := s.Properties.GetOrDefault(config.KeyStatementName, types.GenerateStatementName())
	defer s.Properties.Delete(config.KeyStatementName)

	// Process remote statements
	computePoolId := s.appOptions.GetComputePoolId()
	properties := s.Properties.GetNonLocalProperties()

	var principal string
	serviceAccount := s.Properties.Get(config.KeyServiceAccount)
	if serviceAccount != "" {
		principal = serviceAccount
	} else {
		principal = s.appOptions.GetContext().GetUser().GetResourceId()
	}

	statementObj, err := s.authenticatedGatewayClient().CreateStatement(
		createSqlV1Statement(statement, statementName, computePoolId, properties),
		principal,
		s.appOptions.GetEnvironmentId(),
		s.appOptions.GetOrganizationId(),
	)
	if err != nil {
		status := statementObj.GetStatus()
		return nil, types.NewStatementErrorFailureMsg(err, status.GetDetail())
	}
	return types.NewProcessedStatement(statementObj), nil
}

func createSqlV1Statement(statement string, statementName string, computePoolId string, properties map[string]string) flinkgatewayv1.SqlV1Statement {
	return flinkgatewayv1.SqlV1Statement{
		Name: &statementName,
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Statement:     &statement,
			ComputePoolId: &computePoolId,
			Properties:    &properties,
		},
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
				StatusCode:     types.StatusCode(err),
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
	statementResultObj, err := s.authenticatedGatewayClient().GetStatementResults(s.appOptions.GetEnvironmentId(), statement.StatementName, s.appOptions.GetOrganizationId(), statement.PageToken)
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}

	statementResults := statementResultObj.GetResults()
	convertedResults, err := results.ConvertToInternalResults(statementResults.GetData(), statement.Traits.GetSchema())
	if err != nil {
		return nil, types.NewStatementError(err)
	}
	statement.StatementResults = convertedResults

	statementMetadata := statementResultObj.GetMetadata()
	extractedToken, err := extractPageToken(statementMetadata.GetNext())
	if err != nil {
		return nil, types.NewStatementError(err)
	}
	statement.PageToken = extractedToken
	return &statement, nil
}

func (s *Store) DeleteStatement(statementName string) bool {
	if err := s.authenticatedGatewayClient().DeleteStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrganizationId()); err != nil {
		log.CliLogger.Warnf("Failed to delete the statement: %v", err)
		return false
	}
	log.CliLogger.Infof("Successfully deleted statement: %s", statementName)
	return true
}

func (s *Store) StopStatement(statementName string) bool {
	statement, err := s.authenticatedGatewayClient().GetStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrganizationId())

	if err != nil {
		log.CliLogger.Warnf("Failed to fetch statement to stop it: %v", err)
		return false
	}

	spec, isSpecOk := statement.GetSpecOk()
	if !isSpecOk {
		log.CliLogger.Warnf("Spec for statement that should be stopped is nil")
		return false
	}
	spec.SetStopped(true)

	if err := s.authenticatedGatewayClient().UpdateStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrganizationId(), statement); err != nil {
		log.CliLogger.Warnf("Failed to stop the statement: %v", err)
		return false
	}

	log.CliLogger.Infof("Successfully stopped statement: %s", statementName)
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
			s.DeleteStatement(statementName)
			return nil, &types.StatementError{Message: "result retrieval aborted. Statement will be deleted", StatusCode: 499}
		default:
			start := time.Now()
			statementObj, err := s.authenticatedGatewayClient().GetStatement(s.appOptions.GetEnvironmentId(), statementName, s.appOptions.GetOrganizationId())
			getRequestDuration = time.Since(start)

			statusDetail := s.getStatusDetail(statementObj)
			if err != nil {
				return nil, types.NewStatementErrorFailureMsg(err, statusDetail)
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
			output.Printf(false, "Waiting for statement to be ready. Statement phase is %s. (Timeout %ds/%ds) \n", phase, int(elapsedWaitTime.Seconds()), int(timeout.Seconds()))
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
		Message: fmt.Sprintf("statement is still pending after %f seconds. If you want to increase the timeout for the client, you can run \"SET '%s'='10000';\" to adjust the maximum timeout in milliseconds.",
			timeout.Seconds(), config.KeyResultsTimeout),
		FailureMessage: errorsMsg,
	}
}

func (s *Store) getStatusDetail(statementObj flinkgatewayv1.SqlV1Statement) string {
	status := statementObj.GetStatus()
	if status.GetDetail() != "" {
		return status.GetDetail()
	}

	// if the status detail field is empty, we check if there's an exception instead
	exceptions, err := s.authenticatedGatewayClient().GetExceptions(s.appOptions.GetEnvironmentId(), statementObj.GetName(), s.appOptions.GetOrganizationId())
	if err != nil {
		return ""
	}
	if len(exceptions) < 1 {
		return ""
	}

	// most recent exception is on top of the returned list
	return exceptions[0].GetMessage()
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
		Properties:       NewUserProperties(getDefaultProperties(appOptions), getInitialProperties(appOptions)),
		client:           client,
		exitApplication:  exitApplication,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}
}

func getDefaultProperties(appOptions *types.ApplicationOptions) map[string]string {
	properties := map[string]string{
		config.KeyServiceAccount: appOptions.GetServiceAccountId(),
		config.KeyLocalTimeZone:  getLocalTimezone(),
	}

	return properties
}

func getInitialProperties(appOptions *types.ApplicationOptions) map[string]string {
	properties := map[string]string{}

	if appOptions.GetEnvironmentName() != "" {
		properties[config.KeyCatalog] = appOptions.GetEnvironmentName()
	}
	if appOptions.GetDatabase() != "" {
		properties[config.KeyDatabase] = appOptions.GetDatabase()
	}

	return properties
}

func (s *Store) WaitForTerminalStatementState(ctx context.Context, statement types.ProcessedStatement) (*types.ProcessedStatement, *types.StatementError) {
	for !statement.IsTerminalState() {
		select {
		case <-ctx.Done():
			output.Println(false, "Detached from statement.")
			return &statement, nil
		default:
			statementObj, err := s.authenticatedGatewayClient().GetStatement(s.appOptions.GetEnvironmentId(), statement.StatementName, s.appOptions.GetOrganizationId())
			status := statementObj.GetStatus()
			statusDetail := status.GetDetail()
			if err != nil {
				return nil, &types.StatementError{
					Message:        err.Error(),
					FailureMessage: statusDetail,
					StatusCode:     types.StatusCode(err),
				}
			}

			statement.Status = types.PHASE(statementObj.Status.GetPhase())
			statement.StatusDetail = statusDetail
			if statement.IsTerminalState() {
				break
			}

			if statusDetail != "" {
				output.Println(false, statusDetail)
			}

			time.Sleep(time.Second)
		}
	}

	return &statement, nil
}
