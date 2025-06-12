package store

import (
	"context"
	_ "embed"
	"fmt"
	"strings"
	"time"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/results"
	"github.com/confluentinc/cli/v4/pkg/flink/internal/utils"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type StoreOnPrem struct {
	Properties       types.UserPropertiesInterface
	exitApplication  func()
	client           *flink.CmfRestClient
	appOptions       *types.ApplicationOptions
	tokenRefreshFunc func() error
}

func (s *StoreOnPrem) authenticatedCmfClient() *flink.CmfRestClient {
	if authErr := s.tokenRefreshFunc(); authErr != nil {
		log.CliLogger.Warnf("Failed to refresh token: %v", authErr)
	}
	return s.client
}

func (s *StoreOnPrem) ProcessLocalStatement(statement string) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	defer s.persistUserProperties()
	switch statementType := parseStatementType(statement); statementType {
	case SetStatement:
		return s.processSetStatement(statement)
	case ResetStatement:
		return s.processResetStatement(statement)
	case UseStatement:
		return s.processUseStatement(statement)
	case QuitStatement, ExitStatement:
		s.exitApplication()
		return nil, nil
	default:
		return nil, nil
	}
}

func (s *StoreOnPrem) persistUserProperties() {
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

func (s *StoreOnPrem) ProcessStatement(statement string) (*types.ProcessedStatementOnPrem, *types.StatementError) {
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

	utils.OutputInfof("Statement name: %s\nSubmitting statement...", statementName)
	statementObj, err := s.authenticatedCmfClient().CreateStatement(
		context.Background(),
		s.appOptions.GetEnvironmentId(),
		createSqlV1StatementOnPrem(statement, statementName, computePoolId, properties),
	)

	if err != nil {
		status := statementObj.GetStatus()
		return nil, types.NewStatementErrorFailureMsg(err, status.GetDetail())
	}
	return types.NewProcessedStatementOnPrem(statementObj), nil
}

// TODO: having the Flink configuration can be hard, ignore it for now
// TODO: also check how the properties are passed
func createSqlV1StatementOnPrem(statement string, statementName string, computePoolName string, properties map[string]string) cmfsdk.Statement {
	return cmfsdk.Statement{
		ApiVersion: "cmf.confluent.io/v1",
		Kind:       "Statement",
		Metadata: cmfsdk.StatementMetadata{
			Name: statementName,
		},
		Spec: cmfsdk.StatementSpec{
			Statement:       statement,
			Properties:      &properties,
			ComputePoolName: computePoolName,
			Parallelism:     cmfsdk.PtrInt32(int32(1)),
			Stopped:         cmfsdk.PtrBool(false),
		},
	}
}

func (s *StoreOnPrem) WaitPendingStatement(ctx context.Context, statement types.ProcessedStatementOnPrem) (*types.ProcessedStatementOnPrem, *types.StatementError) {
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

// FetchStatementResults TODO: need to revisit this function on how should we present the SQL statement
func (s *StoreOnPrem) FetchStatementResults(statement types.ProcessedStatementOnPrem) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	// Process local statements
	if statement.IsLocalStatement {
		return &statement, nil
	}

	// Process remote statements that are now running or completed
	statementResultObj, err := s.authenticatedCmfClient().GetStatementResults(context.Background(), s.appOptions.GetEnvironmentId(), statement.StatementName, "")
	if err != nil {
		return nil, &types.StatementError{Message: err.Error()}
	}

	statementResults := statementResultObj.GetResults()
	convertedResults, err := results.ConvertToInternalResultsOnPrem(statementResults, statement.Traits.GetSchema())
	if err != nil {
		return nil, types.NewStatementError(err)
	}
	statement.StatementResults = convertedResults

	//TODO: we have CCloud page token here, CMF SDK does not support it yet, double check with Fabian
	return &statement, nil
}

func (s *StoreOnPrem) DeleteStatement(statementName string) bool {
	// TODO: check if the context should be propagated from the caller
	if err := s.authenticatedCmfClient().DeleteStatement(context.Background(), s.appOptions.EnvironmentName, statementName); err != nil {
		log.CliLogger.Warnf("Failed to delete the statement: %v", err)
		return false
	}
	log.CliLogger.Infof("Successfully deleted statement: %s", statementName)
	return true
}

func (s *StoreOnPrem) StopStatement(statementName string) bool {
	// TODO: check if the context should be propagated from the caller
	statement, err := s.authenticatedCmfClient().GetStatement(context.Background(), s.appOptions.EnvironmentName, statementName)

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

	if err := s.authenticatedCmfClient().UpdateStatement(context.Background(), statementName, s.appOptions.GetEnvironmentName(), statement); err != nil {
		log.CliLogger.Warnf("Failed to stop the statement: %v", err)
		return false
	}

	log.CliLogger.Infof("Successfully stopped statement: %s", statementName)
	return true
}

func (s *StoreOnPrem) waitForPendingStatement(ctx context.Context, statementName string, timeout time.Duration) (*types.ProcessedStatementOnPrem, *types.StatementError) {
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
			statementObj, err := s.authenticatedCmfClient().GetStatement(context.Background(), s.appOptions.GetEnvironmentId(), statementName)
			getRequestDuration = time.Since(start)

			statusDetail := s.getStatusDetail(statementObj)
			if err != nil {
				return nil, types.NewStatementErrorFailureMsg(err, statusDetail)
			}

			phase = types.PHASE(statementObj.Status.GetPhase())
			if phase != types.PENDING {
				processedStatement := types.NewProcessedStatementOnPrem(statementObj)
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

func (s *StoreOnPrem) getStatusDetail(statementObj cmfsdk.Statement) string {
	status := statementObj.GetStatus()
	if status.GetDetail() != "" {
		return status.GetDetail()
	}

	// if the status detail field is empty, we check if there's an exception instead
	exceptionList, err := s.authenticatedCmfClient().ListStatementExceptions(context.Background(), s.appOptions.GetEnvironmentId(), statementObj.Metadata.GetName())
	if err != nil {
		return ""
	}
	exceptions := exceptionList.GetData()
	if len(exceptions) < 1 {
		return ""
	}

	// most recent exception is on top of the returned list
	topException := &exceptions[0]
	return topException.GetMessage()
}

func NewStoreOnPrem(client *flink.CmfRestClient, exitApplication func(), userProperties types.UserPropertiesInterface, appOptions *types.ApplicationOptions, tokenRefreshFunc func() error) types.StoreInterfaceOnPrem {
	return &StoreOnPrem{
		Properties:       userProperties,
		client:           client,
		exitApplication:  exitApplication,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}
}

func (s *StoreOnPrem) WaitForTerminalStatementState(ctx context.Context, statement types.ProcessedStatementOnPrem) (*types.ProcessedStatementOnPrem, *types.StatementError) {
	for !statement.IsTerminalState() {
		select {
		case <-ctx.Done():
			output.Println(false, "Detached from statement.")
			return &statement, nil
		default:
			statementObj, err := s.authenticatedCmfClient().GetStatement(context.Background(), s.appOptions.GetEnvironmentId(), statement.StatementName)
			status := statementObj.GetStatus()
			statusDetail := status.GetDetail()
			if err != nil {
				return nil, &types.StatementError{
					Message:        err.Error(),
					FailureMessage: statusDetail,
					StatusCode:     types.StatusCode(err),
				}
			}

			statement.Status = types.PHASE(status.GetPhase())
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
