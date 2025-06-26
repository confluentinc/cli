package store

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors/flink"
	pflink "github.com/confluentinc/cli/v4/pkg/flink"
	flinkconfig "github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/test"
	"github.com/confluentinc/cli/v4/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v4/pkg/flink/types"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

type StoreTestSuite struct {
	suite.Suite
}

const (
	testStatementName       = "statement-name"
	testStatusDetailMessage = "Test status detail message"
	selectFromStatement     = "SELECT * FROM table"
)

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func tokenRefreshFunc() error {
	return nil
}

func (s *StoreTestSuite) TestGenerateStatementName() {
	statementRegex := `^cli-\d{4}-\d{2}-\d{2}-\d{6}-[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`
	for i := 0; i < 10; i++ {
		s.Require().Regexp(statementRegex, types.GenerateStatementName())
	}
}

func TestStoreProcessLocalStatement(t *testing.T) {
	// Create new stores
	stores := make([]types.StoreInterface, 2)

	// Cloud store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	mockAppController := mock.NewMockApplicationControllerInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	userProperties := NewUserProperties(&appOptions)
	stores[0] = NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc).(*Store)

	// On-prem store
	cmfClient, err := pflink.NewCmfRestClient(cmfsdk.NewConfiguration(), &pflink.OnPremCMFRestFlagValues{}, true)
	require.NoError(t, err)
	userProperties = NewUserProperties(&appOptions)
	stores[1] = NewStoreOnPrem(cmfClient, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc).(*StoreOnPrem)

	for _, s := range stores {
		result, err := s.ProcessLocalStatement("SET 'foo'='bar';")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsLocalStatement)

		result, err = s.ProcessLocalStatement("RESET;")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsLocalStatement)

		result, err = s.ProcessLocalStatement("USE CATALOG my_catalog;")
		assert.Nil(t, err)
		assert.NotNil(t, result)
		assert.True(t, result.IsLocalStatement)

		result, err = s.ProcessLocalStatement("SELECT * FROM users;")
		assert.Nil(t, err)
		assert.Nil(t, result)

		mockAppController.EXPECT().ExitApplication()
		result, err = s.ProcessLocalStatement("EXIT;")
		assert.Nil(t, err)
		assert.Nil(t, result)

		mockAppController.EXPECT().ExitApplication()
		result, err = s.ProcessLocalStatement("quit")
		assert.Nil(t, err)
		assert.Nil(t, result)
	}
}

func TestWaitForPendingCompletedStatement(t *testing.T) {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil)

		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, time.Duration(10))
		assert.Nil(t, err)
		assert.NotNil(t, processedStatement)
		assert.Equal(t, types.NewProcessedStatement(statementObj), processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		s := &StoreOnPrem{
			client:           client,
			appOptions:       &types.ApplicationOptions{EnvironmentId: "envId"},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "COMPLETED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil)

		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, time.Duration(10))
		assert.Nil(t, err)
		assert.NotNil(t, processedStatement)
		assert.Equal(t, types.NewProcessedStatementOnPrem(statementObj), processedStatement)
	}
}

func TestWaitForPendingTimesOut(t *testing.T) {
	timeout := time.Duration(10) * time.Millisecond
	statementErrorMessage := fmt.Sprintf("statement is still pending after %f seconds. If you want to increase the timeout for the client, you can run \"SET '%s'='10000';\" to adjust the maximum timeout in milliseconds.", timeout.Seconds(), flinkconfig.KeyResultsTimeout)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "PENDING",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		expectedError := &types.StatementError{
			Message:        statementErrorMessage,
			FailureMessage: fmt.Sprintf("captured retryable errors: %s", testStatusDetailMessage),
		}
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil).AnyTimes()
		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, timeout)

		assert.Equal(t, expectedError, err)
		assert.Nil(t, processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		s := &StoreOnPrem{
			client:           client,
			appOptions:       &types.ApplicationOptions{EnvironmentId: "envId"},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "PENDING",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}
		expectedError := &types.StatementError{
			Message: statementErrorMessage,
		}
		client.EXPECT().CmfApiContext().Return(context.Background()).AnyTimes()
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil).AnyTimes()
		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, timeout)

		assert.Equal(t, expectedError, err)
		assert.Nil(t, processedStatement)
	}
}

// Cloud only; On-prem does not have retryable errors
func TestWaitForPendingHitsErrorRetryLimit(t *testing.T) {
	timeout := 10 * time.Second

	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementObj := flinkgatewayv1.SqlV1Statement{
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase:  "PENDING",
			Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
		},
	}
	expectedError := &types.StatementError{
		Message:        "the server can't process this statement right now, exiting after 6 retries",
		FailureMessage: fmt.Sprintf("captured retryable errors: %s", strings.Repeat(testStatusDetailMessage+"; ", 5)+testStatusDetailMessage),
	}
	client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil).AnyTimes()
	processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, timeout)

	assert.Equal(t, expectedError, err)
	assert.Nil(t, processedStatement)
}

func TestWaitForPendingEventuallyCompletes(t *testing.T) {
	transientStatusDetailMessage := "Transient status detail message"
	finalStatusDetailMessage := "Final status detail message"

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "PENDING",
				Detail: &transientStatusDetailMessage,
			},
		}

		statementObjCompleted := flinkgatewayv1.SqlV1Statement{
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: &finalStatusDetailMessage,
			},
		}
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil).Times(3)
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObjCompleted, nil)

		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, time.Duration(10)*time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, processedStatement)
		assert.Equal(t, types.NewProcessedStatement(statementObjCompleted), processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		s := &StoreOnPrem{
			client:           client,
			appOptions:       &types.ApplicationOptions{EnvironmentId: "envId"},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "PENDING",
				Detail: cmfsdk.PtrString(transientStatusDetailMessage),
			},
		}

		statementObjCompleted := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "COMPLETED",
				Detail: cmfsdk.PtrString(finalStatusDetailMessage),
			},
		}
		client.EXPECT().CmfApiContext().Return(context.Background()).Times(4)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil).Times(3)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObjCompleted, nil)

		processedStatement, err := s.waitForPendingStatement(context.Background(), testStatementName, time.Duration(10)*time.Second)
		assert.Nil(t, err)
		assert.NotNil(t, processedStatement)
		assert.Equal(t, types.NewProcessedStatementOnPrem(statementObjCompleted), processedStatement)
	}
}

func TestWaitForPendingStatementErrors(t *testing.T) {
	waitTime := time.Millisecond * 1
	returnedError := flink.NewError("couldn't get statement", "", http.StatusInternalServerError)
	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: testStatusDetailMessage,
		StatusCode:     http.StatusInternalServerError,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}
		statementObj := flinkgatewayv1.SqlV1Statement{
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}

		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, returnedError)
		_, err := s.waitForPendingStatement(context.Background(), testStatementName, waitTime)
		assert.Equal(t, expectedError, err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		s := &StoreOnPrem{
			client:           client,
			appOptions:       &types.ApplicationOptions{EnvironmentId: "envId"},
			tokenRefreshFunc: tokenRefreshFunc,
		}
		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "COMPLETED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background()).AnyTimes()
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, returnedError)
		_, err := s.waitForPendingStatement(context.Background(), testStatementName, waitTime)
		assert.Equal(t, expectedError, err)
	}
}

func TestCancelPendingStatement(t *testing.T) {
	waitTime := time.Second * 1
	ctx, cancelFunc := context.WithCancel(context.Background())
	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	expectedErr := &types.StatementError{Message: "result retrieval aborted. Statement will be deleted", StatusCode: http.StatusInternalServerError}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase: "PENDING",
			},
		}

		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil).AnyTimes()
		client.EXPECT().DeleteStatement("envId", testStatementName, "orgId").Return(nil).AnyTimes()
		client.EXPECT().GetExceptions("envId", testStatementName, "orgId").Return([]flinkgatewayv1.SqlV1StatementException{}, flinkError).AnyTimes()

		// Schedule routine to cancel context
		go func() {
			time.Sleep(time.Millisecond * 20)
			cancelFunc()
		}()

		res, err := s.waitForPendingStatement(ctx, testStatementName, waitTime)
		assert.Nil(t, res)
		assert.EqualError(t, err, expectedErr.Error())
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		s := &StoreOnPrem{
			client:           client,
			appOptions:       &types.ApplicationOptions{EnvironmentId: "envId"},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase: "PENDING",
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background()).AnyTimes()
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil).AnyTimes()
		client.EXPECT().DeleteStatement(context.Background(), "envId", testStatementName).Return(nil).AnyTimes()
		client.EXPECT().ListStatementExceptions(context.Background(), "envId", testStatementName).Return(cmfsdk.StatementExceptionList{}, flinkError).AnyTimes()

		// Schedule routine to cancel context
		go func() {
			time.Sleep(time.Millisecond * 20)
			cancelFunc()
		}()

		res, err := s.waitForPendingStatement(ctx, testStatementName, waitTime)
		assert.Nil(t, res)
		assert.EqualError(t, err, expectedErr.Error())
	}
}

func (s *StoreTestSuite) TestIsSetStatement() {
	assert.True(s.T(), true, statementStartsWithOp("SET", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key=value", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET   ", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("    set   ", flinkconfig.OpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", flinkconfig.OpSet))

	assert.False(s.T(), false, statementStartsWithOp("SETting", flinkconfig.OpSet))
	assert.False(s.T(), false, statementStartsWithOp("", flinkconfig.OpSet))
	assert.False(s.T(), false, statementStartsWithOp("should be false", flinkconfig.OpSet))
	assert.False(s.T(), false, statementStartsWithOp("USE", flinkconfig.OpSet))
	assert.False(s.T(), false, statementStartsWithOp("SETTING", flinkconfig.OpSet))
}

func (s *StoreTestSuite) TestIsUseStatement() {
	assert.True(s.T(), statementStartsWithOp("USE", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("USE catalog", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("use CATALOG cat", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("USE   ", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("use   ", flinkconfig.OpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", flinkconfig.OpUse))

	assert.False(s.T(), statementStartsWithOp("SET", flinkconfig.OpUse))
	assert.False(s.T(), statementStartsWithOp("USES", flinkconfig.OpUse))
	assert.False(s.T(), statementStartsWithOp("", flinkconfig.OpUse))
	assert.False(s.T(), statementStartsWithOp("should be false", flinkconfig.OpUse))
}

func (s *StoreTestSuite) TestIsResetStatement() {
	assert.True(s.T(), true, statementStartsWithOp("RESET", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET   ", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("reset   ", flinkconfig.OpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", flinkconfig.OpReset))

	assert.False(s.T(), false, statementStartsWithOp("RESETting", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("USE", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("RESETTING", flinkconfig.OpReset))
}

func (s *StoreTestSuite) TestIsExitStatement() {
	assert.True(s.T(), true, statementStartsWithOp("EXIT", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("EXIT ;", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit   ;", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("exiT   ", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("Exit   ", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("eXit   ", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit", flinkconfig.OpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit ", flinkconfig.OpExit))

	assert.False(s.T(), false, statementStartsWithOp("exits", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("exitt;", flinkconfig.OpReset))
	assert.False(s.T(), false, statementStartsWithOp("exi", flinkconfig.OpReset))
}

func (s *StoreTestSuite) TestIsQuitStatement() {
	assert.True(s.T(), true, statementStartsWithOp("QUIT", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("QUIT ;", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("quit   ;", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("quiT   ", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("Quit   ", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("qUit   ", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("quit", flinkconfig.OpQuit))
	assert.True(s.T(), true, statementStartsWithOp("quit ", flinkconfig.OpQuit))

	assert.False(s.T(), false, statementStartsWithOp("quits", flinkconfig.OpQuit))
	assert.False(s.T(), false, statementStartsWithOp("", flinkconfig.OpQuit))
	assert.False(s.T(), false, statementStartsWithOp("should be false", flinkconfig.OpQuit))
	assert.False(s.T(), false, statementStartsWithOp("quitt;", flinkconfig.OpQuit))
	assert.False(s.T(), false, statementStartsWithOp("qui", flinkconfig.OpQuit))
}

func (s *StoreTestSuite) TestParseSetStatement() {
	key, value, err := parseSetStatement("SET 'key'='value'")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("SET 'key'='value';")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key'='value'    ;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key' = 'value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key'     =    'value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key    '= '		va  lue'    ")
	assert.Equal(s.T(), "key    ", key)
	assert.Equal(s.T(), "\t\tva  lue", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key' ='value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set 'key'		 ='value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("SET")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("sET 	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("sET 'key'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement("sET = 'value'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement("sET 'key'= \n'value'	")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement("set key= \nvalue	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement("set 'key'= \nvalue	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement("set key= \n'value'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement("set 'key= \nvalue'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement(`set ''key''=''value''`)
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement(`set '''key'''='''value'''`)
	assert.Equal(s.T(), "'key'", key)
	assert.Equal(s.T(), "'value'", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement(`set ''''key'''='''value'''`)
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
	assert.Error(s.T(), err)

	key, value, err = parseSetStatement(`set 'key'''='value'`)
	assert.Equal(s.T(), "key'", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)

	key, value, err = parseSetStatement(`set 'key'''''='value'`)
	assert.Equal(s.T(), "key''", key)
	assert.Equal(s.T(), "value", value)
	assert.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestParseSetStatementError() {
	_, _, err := parseSetStatement("SET key")
	assert.Equal(s.T(), &types.StatementError{
		Message: `missing "="`,
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement("SET =")
	assert.Equal(s.T(), &types.StatementError{
		Message: "key and value not present",
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement("SET key=")
	assert.Equal(s.T(), &types.StatementError{
		Message:    "value for key not present",
		Suggestion: `if you want to reset a key, use "RESET 'key'"`,
	}, err)

	_, _, err = parseSetStatement("SET =value")
	assert.Equal(s.T(), &types.StatementError{
		Message: "key not present",
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement("SET key=value=as")
	assert.Equal(s.T(), &types.StatementError{
		Message: `key and value must be enclosed by single quotes (')`,
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement("SET key=value")
	assert.Equal(s.T(), &types.StatementError{
		Message: "key and value must be enclosed by single quotes (')",
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement(`set ''key'''=''value''`)
	assert.Equal(s.T(), &types.StatementError{
		Message:    "key or value contains unescaped single quotes (')",
		Usage:      []string{"SET 'key'='value'"},
		Suggestion: `please escape all single quotes with another single quote "''key''"`,
	}, err)

	_, _, err = parseSetStatement(`set 'key'=''value''`)
	assert.Equal(s.T(), &types.StatementError{
		Message:    "key or value contains unescaped single quotes (')",
		Usage:      []string{"SET 'key'='value'"},
		Suggestion: `please escape all single quotes with another single quote "''key''"`,
	}, err)
}

func (s *StoreTestSuite) TestParseUseStatement() {
	catalog, database, _ := parseUseStatement("USE CATALOG c;")
	assert.Equal(s.T(), "c", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("use   catalog   \nc   ")
	assert.Equal(s.T(), "c", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("use   catalog     ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("catalog   c")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("use     db   ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "db", database)

	catalog, database, _ = parseUseStatement("dAtaBaSe  db   ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("use     \ndatabase_name   ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "database_name", database)
}

func (s *StoreTestSuite) TestParseUseStatementCatalogPath() {
	catalog, database, _ := parseUseStatement("USE CATALOG `my catalog-123`")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "my catalog-123", catalog)

	catalog, database, _ = parseUseStatement("USE CATALOG `cat`")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "cat", catalog)

	catalog, database, _ = parseUseStatement("use catalog `cAt`")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "cAt", catalog)

	catalog, database, _ = parseUseStatement("USE CATALOG `ca``t`")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "ca`t", catalog)

	catalog, database, _ = parseUseStatement("use   catalog   \n`cat`   ")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "cat", catalog)

	catalog, database, _ = parseUseStatement("use   catalog     ")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "", catalog)

	catalog, database, _ = parseUseStatement("catalog   `c`")
	assert.Equal(s.T(), "", database)
	assert.Equal(s.T(), "", catalog)
}

func (s *StoreTestSuite) TestParseUseStatementDatabasePath() {
	catalog, database, _ := parseUseStatement("USE `my db-123`")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "my db-123", database)

	catalog, database, _ = parseUseStatement("USE `db`")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "db", database)

	catalog, database, _ = parseUseStatement("use `dB`")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "dB", database)

	catalog, database, _ = parseUseStatement("USE `d``B`")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "d`B", database)

	catalog, database, _ = parseUseStatement("use     \n`db`   ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "db", database)

	catalog, database, _ = parseUseStatement("use        ")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "", database)

	catalog, database, _ = parseUseStatement("  `c`")
	assert.Equal(s.T(), "", catalog)
	assert.Equal(s.T(), "", database)
}

func (s *StoreTestSuite) TestParseUseStatementCatalogDatabasePath() {
	catalog, database, _ := parseUseStatement("USE `my catalog`.`my database`")
	assert.Equal(s.T(), "my catalog", catalog)
	assert.Equal(s.T(), "my database", database)

	catalog, database, _ = parseUseStatement("USE `my catalog`.`my_database`")
	assert.Equal(s.T(), "my catalog", catalog)
	assert.Equal(s.T(), "my_database", database)

	catalog, database, _ = parseUseStatement("USE `my_catalog`.`my database`")
	assert.Equal(s.T(), "my_catalog", catalog)
	assert.Equal(s.T(), "my database", database)

	catalog, database, _ = parseUseStatement("USE `my_catalog`.`my_database`")
	assert.Equal(s.T(), "my_catalog", catalog)
	assert.Equal(s.T(), "my_database", database)

	catalog, database, _ = parseUseStatement("USE `my catalog`.database")
	assert.Equal(s.T(), "my catalog", catalog)
	assert.Equal(s.T(), "database", database)

	catalog, database, _ = parseUseStatement("USE cat.`my database`")
	assert.Equal(s.T(), "cat", catalog)
	assert.Equal(s.T(), "my database", database)

	catalog, database, _ = parseUseStatement("USE `my catalog`   .   `my database`")
	assert.Equal(s.T(), "my catalog", catalog)
	assert.Equal(s.T(), "my database", database)

	catalog, database, _ = parseUseStatement("USE cat   .   db")
	assert.Equal(s.T(), "cat", catalog)
	assert.Equal(s.T(), "db", database)

	catalog, database, _ = parseUseStatement("USE cat.   db")
	assert.Equal(s.T(), "cat", catalog)
	assert.Equal(s.T(), "db", database)

	catalog, database, _ = parseUseStatement("USE cat   .db")
	assert.Equal(s.T(), "cat", catalog)
	assert.Equal(s.T(), "db", database)
}

func (s *StoreTestSuite) TestParseUseStatementError() {
	_, _, err := parseUseStatement("USE CATALOG ;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax for USE CATALOG\nUsage: \"USE CATALOG `my_catalog`\"", err.Error())

	_, _, err = parseUseStatement("USE;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax for USE\nUsage: \"USE CATALOG `my_catalog`\", \"USE `my_database`\", or \"USE `my_catalog`.`my_database`\"", err.Error())

	_, _, err = parseUseStatement("USE CATALOG DATABASE DB2;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax for USE CATALOG\nUsage: \"USE CATALOG `my_catalog`\"", err.Error())

	_, _, err = parseUseStatement("USE `use`.`CATALOG`.`table` ;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax for USE\nUsage: \"USE CATALOG `my_catalog`\", \"USE `my_database`\", or \"USE `my_catalog`.`my_database`\"", err.Error())
}

func (s *StoreTestSuite) TestParseResetStatement() {
	key, err := parseResetStatement("RESET 'key'")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("RESET 'key.key';")
	assert.Equal(s.T(), "key.key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("RESET 'KEY.key';")
	assert.Equal(s.T(), "KEY.key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset 'key'    ;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset 'key'   ")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset 'key';;;;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("RESET")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reSET 	")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reSET 'key'	")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("resET 'KEY' ")
	assert.Equal(s.T(), "KEY", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("resET 'key';;;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset key;")
	assert.Equal(s.T(), "", key)
	assert.Error(s.T(), err)

	key, err = parseResetStatement("reset key';")
	assert.Equal(s.T(), "", key)
	assert.Error(s.T(), err)

	key, err = parseResetStatement("reset 'key;")
	assert.Equal(s.T(), "", key)
	assert.Error(s.T(), err)

	key, err = parseResetStatement("reset 'key one';")
	assert.Equal(s.T(), "key one", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("reset ''key one'';")
	assert.Equal(s.T(), "", key)
	assert.Error(s.T(), err)

	key, err = parseResetStatement(`reset '''key one''';`)
	assert.Equal(s.T(), "'key one'", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement(`reset ''''key one''';`)
	assert.Equal(s.T(), "", key)
	assert.Error(s.T(), err)

	key, err = parseResetStatement(`reset 'key'' one';`)
	assert.Equal(s.T(), "key' one", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement(`reset 'key'''' one';`)
	assert.Equal(s.T(), "key'' one", key)
	assert.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestParseResetStatementError() {
	key, err := parseResetStatement(" ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("RESET key key2")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("RESET key key2 key3")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("RESET key;; key key3")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("RESET key key;;; key3")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("RESET key;")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message: "invalid syntax for RESET, key must be enclosed by single quotes ''",
		Usage:   []string{"RESET 'key'"},
	}, err)

	key, err = parseResetStatement("reset ''key one'';")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), &types.StatementError{
		Message:    "key contains unescaped single quotes (')",
		Usage:      []string{"RESET 'key'"},
		Suggestion: `please escape all single quotes with another single quote "''key''"`,
	}, err)
}

func (s *StoreTestSuite) TestStopStatement() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		statementObj := flinkgatewayv1.NewSqlV1StatementWithDefaults()
		spec := flinkgatewayv1.NewSqlV1StatementSpecWithDefaults()
		statementObj.SetName(testStatementName)
		statementObj.SetSpec(*spec)
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(*statementObj, nil)

		statementUpdated := flinkgatewayv1.NewSqlV1StatementWithDefaults()
		specUpdated := flinkgatewayv1.NewSqlV1StatementSpecWithDefaults()
		statementUpdated.SetName(testStatementName)
		specUpdated.SetStopped(true)
		statementUpdated.SetSpec(*specUpdated)
		client.EXPECT().UpdateStatement("envId", testStatementName, "orgId", *statementUpdated).Return(nil)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.True(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background()).Times(2)
		statementObj := cmfsdk.NewStatementWithDefaults()
		spec := cmfsdk.NewStatementSpecWithDefaults()
		statementObj.Metadata.SetName(testStatementName)
		statementObj.SetSpec(*spec)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(*statementObj, nil)

		statementUpdated := cmfsdk.NewStatementWithDefaults()
		specUpdated := cmfsdk.NewStatementSpecWithDefaults()
		statementUpdated.Metadata.SetName(testStatementName)
		specUpdated.SetStopped(true)
		statementUpdated.SetSpec(*specUpdated)
		client.EXPECT().UpdateStatement(context.Background(), "envId", testStatementName, *statementUpdated).Return(nil)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.True(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestStopStatementFailsOnGetError() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}

		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(flinkgatewayv1.SqlV1Statement{}, flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(cmfsdk.Statement{}, flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestStopStatementFailsOnNilSpecError() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	statementObj := flinkgatewayv1.NewSqlV1StatementWithDefaults()
	statementObj.SetName(testStatementName)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(flinkgatewayv1.SqlV1Statement{}, flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(cmfsdk.Statement{}, flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestStopStatementFailsOnUpdateError() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)

		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		statementObj := flinkgatewayv1.NewSqlV1StatementWithDefaults()
		spec := flinkgatewayv1.NewSqlV1StatementSpecWithDefaults()
		statementObj.SetName(testStatementName)
		statementObj.SetSpec(*spec)

		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(*statementObj, nil)
		statementObj.Spec.SetStopped(true)
		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().UpdateStatement("envId", testStatementName, "orgId", *statementObj).Return(flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background()).Times(2)
		statementObj := cmfsdk.NewStatementWithDefaults()
		spec := cmfsdk.NewStatementSpecWithDefaults()
		statementObj.Metadata.SetName(testStatementName)
		statementObj.SetSpec(*spec)

		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(*statementObj, nil)
		statementObj.Spec.SetStopped(true)
		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().UpdateStatement(context.Background(), "envId", testStatementName, *statementObj).Return(flinkError)

		wasStatementDeleted := store.StopStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestDeleteStatement() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().DeleteStatement("envId", testStatementName, "orgId").Return(nil)

		wasStatementDeleted := store.DeleteStatement(testStatementName)
		require.True(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().DeleteStatement(context.Background(), "envId", testStatementName).Return(nil)

		wasStatementDeleted := store.DeleteStatement(testStatementName)
		require.True(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestDeleteStatementFailsOnError() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().DeleteStatement("envId", testStatementName, "orgId").Return(flinkError)
		wasStatementDeleted := store.DeleteStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		flinkError := flink.NewError("error", "", http.StatusInternalServerError)
		client.EXPECT().DeleteStatement(context.Background(), "envId", testStatementName).Return(flinkError)
		wasStatementDeleted := store.DeleteStatement(testStatementName)
		require.False(s.T(), wasStatementDeleted)
	}
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWithCompletedStatement() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	statement := types.ProcessedStatement{
		StatementName: testStatementName,
		Status:        types.COMPLETED,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		statementResultObj := flinkgatewayv1.SqlV1StatementResult{
			Metadata: flinkgatewayv1.ResultListMeta{},
			Results:  &flinkgatewayv1.SqlV1StatementResultResults{},
		}
		client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		statementResultObj := cmfsdk.Statement{
			Result: &cmfsdk.StatementResult{},
		}
		client.EXPECT().GetStatement(context.Background(), "envId", statement.StatementName).Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
}

func (s *StoreTestSuite) TestFetchResultsWithRunningStatement() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	statement := types.ProcessedStatement{
		StatementName: testStatementName,
		Status:        types.RUNNING,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		statementResultObj := flinkgatewayv1.SqlV1StatementResult{
			Metadata: flinkgatewayv1.ResultListMeta{},
			Results:  &flinkgatewayv1.SqlV1StatementResultResults{},
		}
		client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		statement.Traits.CmfStatementTraits = &cmfsdk.StatementTraits{
			SqlKind: cmfsdk.PtrString("SELECT"),
		}
		statementResultObj := cmfsdk.StatementResult{
			Results: cmfsdk.StatementResults{},
		}
		client.EXPECT().GetStatementResults(context.Background(), "envId", statement.StatementName, "").Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenPageTokenExists() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	statement := types.ProcessedStatement{
		StatementName: testStatementName,
		Status:        types.RUNNING,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)

		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		nextPage := "https://devel.cpdev.cloud/some/results?page_token=eyJWZX"
		statementResultObj := flinkgatewayv1.SqlV1StatementResult{
			Metadata: flinkgatewayv1.ResultListMeta{Next: &nextPage},
			Results:  &flinkgatewayv1.SqlV1StatementResultResults{},
		}
		client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background()).Times(2)
		statement.Traits.CmfStatementTraits = &cmfsdk.StatementTraits{
			SqlKind: cmfsdk.PtrString("SELECT"),
		}
		statementResultObj := cmfsdk.StatementResult{
			Metadata: cmfsdk.StatementResultMetadata{
				Annotations: &map[string]string{"nextPageToken": "eyJWZX"},
			},
			Results: cmfsdk.StatementResults{},
		}
		client.EXPECT().GetStatementResults(context.Background(), "envId", statement.StatementName, "").Return(statementResultObj, nil)

		firstPageResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), firstPageResults)
		require.Nil(s.T(), err)
		require.Equal(s.T(), "eyJWZX", firstPageResults.PageToken)

		statementResultObj.Metadata.SetAnnotations(nil)
		client.EXPECT().GetStatementResults(context.Background(), "envId", statement.StatementName, "eyJWZX").Return(statementResultObj, nil)

		secondPageResults, err := store.FetchStatementResults(*firstPageResults)
		require.NotNil(s.T(), secondPageResults)
		require.Nil(s.T(), err)
		require.Equal(s.T(), "", secondPageResults.PageToken)
	}
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenResultsExist() {
	ctrl := gomock.NewController(s.T())
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	statement := types.ProcessedStatement{
		StatementName: testStatementName,
		Status:        types.RUNNING,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			OrganizationId:  "orgId",
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStore(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		statementResultObj := flinkgatewayv1.SqlV1StatementResult{
			Metadata: flinkgatewayv1.ResultListMeta{},
			Results:  &flinkgatewayv1.SqlV1StatementResultResults{Data: &[]any{map[string]any{"op": 0}}},
		}
		client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(ctrl)
		appOptions := types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		}
		userProperties := NewUserProperties(&appOptions)
		store := NewStoreOnPrem(client, mockAppController.ExitApplication, userProperties, &appOptions, tokenRefreshFunc)

		client.EXPECT().CmfApiContext().Return(context.Background())
		statement.Traits.CmfStatementTraits = &cmfsdk.StatementTraits{
			SqlKind: cmfsdk.PtrString("SELECT"),
		}
		statementResultObj := cmfsdk.StatementResult{
			Results: cmfsdk.StatementResults{Data: &[]map[string]interface{}{{"op": 0}}},
		}
		client.EXPECT().GetStatementResults(context.Background(), "envId", statement.StatementName, "").Return(statementResultObj, nil)

		statementResults, err := store.FetchStatementResults(statement)
		require.NotNil(s.T(), statementResults)
		require.Nil(s.T(), err)
	}
}

func (s *StoreTestSuite) TestExtractPageToken() {
	token, err := extractPageToken("https://devel.cpdev.cloud/some/results?page_token=eyJWZX")
	require.Equal(s.T(), "eyJWZX", token)
	require.Nil(s.T(), err)
}

func TestCalcWaitTime(t *testing.T) {
	// Define test cases
	testCases := []struct {
		retries          int
		expectedWaitTime time.Duration
	}{
		{0, flinkconfig.InitialWaitTime},
		{3, flinkconfig.InitialWaitTime + time.Duration(flinkconfig.WaitTimeIncrease*0)*time.Millisecond},
		{7, flinkconfig.InitialWaitTime + time.Duration(flinkconfig.WaitTimeIncrease*0)*time.Millisecond},
		{10, flinkconfig.InitialWaitTime + time.Duration(flinkconfig.WaitTimeIncrease*1)*time.Millisecond},
		{15, flinkconfig.InitialWaitTime + time.Duration(flinkconfig.WaitTimeIncrease*1)*time.Millisecond},
		{32, flinkconfig.InitialWaitTime + time.Duration(flinkconfig.WaitTimeIncrease*3)*time.Millisecond},
	}

	for _, testCase := range testCases {
		waitTime := calcWaitTime(testCase.retries)
		require.Equal(t, testCase.expectedWaitTime, waitTime, fmt.Sprintf("For retries=%d, expected wait time=%v, but got %v",
			testCase.retries, testCase.expectedWaitTime, waitTime))
	}
}

func TestTimeout(t *testing.T) {
	testCases := []struct {
		name       string
		properties map[string]string
		expected   time.Duration
	}{
		{
			name: "results-timeout property set",
			properties: map[string]string{
				flinkconfig.KeyResultsTimeout: "10000", // timeout in milliseconds
			},
			expected: 10 * time.Second,
		},
		{
			name:       "results-timeout property not set",
			properties: map[string]string{},
			expected:   flinkconfig.DefaultTimeoutDuration,
		},
		{
			name: "invalid results-timeout property",
			properties: map[string]string{
				flinkconfig.KeyResultsTimeout: "abc", // invalid duration
			},
			expected: flinkconfig.DefaultTimeoutDuration,
		},
	}

	// Iterate over test cases and run the function for each input, comparing output to expected value
	for _, tc := range testCases {
		store := Store{Properties: NewUserPropertiesWithDefaults(tc.properties, map[string]string{})}
		result := getTimeout(store.Properties)
		require.Equal(t, tc.expected, result, tc.name)
	}
}

// Cloud only; On-prem does not use service accounts
func (s *StoreTestSuite) TestProcessStatementWithServiceAccount() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
	}
	serviceAccountId := "sa-123"
	store := Store{
		Properties:       NewUserPropertiesWithDefaults(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId, flinkconfig.KeyStatementName: testStatementName}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1.SqlV1Statement{
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase:  "PENDING",
			Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
		},
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     flinkgatewayv1.PtrString(selectFromStatement),
		},
	}

	client.EXPECT().CreateStatement(SqlV1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
}

// Cloud only; On-prem does not set user identity
func (s *StoreTestSuite) TestProcessStatementWithUserIdentity() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))

	user := "u-1234"
	contextState := &config.ContextState{
		Auth: &config.AuthConfig{
			User: &ccloudv1.User{
				ResourceId: user,
				Email:      "test-user@email",
			},
			Organization: testserver.RegularOrg,
		},
		AuthToken:        "eyJ.eyJ.abc",
		AuthRefreshToken: "v1.abc",
	}
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
		Context:        &config.Context{State: contextState, Config: &config.Config{}},
	}
	store := Store{
		Properties: NewUserPropertiesWithDefaults(
			map[string]string{flinkconfig.KeyStatementName: testStatementName}, map[string]string{},
		),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1.SqlV1Statement{
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase:  "PENDING",
			Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
		},
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     flinkgatewayv1.PtrString(selectFromStatement),
		},
	}

	client.EXPECT().CreateStatement(SqlV1StatementMatcher{statementObj}, user, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
}

func (s *StoreTestSuite) TestProcessStatementOnPrem() {
	client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
	store := StoreOnPrem{
		Properties: NewUserPropertiesWithDefaults(
			map[string]string{flinkconfig.KeyStatementName: testStatementName}, map[string]string{},
		),
		client: client,
		appOptions: &types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		},
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := cmfsdk.Statement{
		Status: &cmfsdk.StatementStatus{
			Phase:  "PENDING",
			Detail: cmfsdk.PtrString(testStatusDetailMessage),
		},
		Spec: cmfsdk.StatementSpec{
			Properties: &nonLocalProperties,
			Statement:  selectFromStatement,
		},
	}

	client.EXPECT().CmfApiContext().Return(context.Background())
	client.EXPECT().CreateStatement(context.Background(), "envId", CmfStatementMatcher{statementObj}).Return(statementObj, nil)

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatementOnPrem(statementObj), processedStatement)
}

func (s *StoreTestSuite) TestProcessStatementFailsOnError() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	serviceAccountId := "serviceAccountId"
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
	}
	store := Store{
		Properties:       NewUserPropertiesWithDefaults(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId, flinkconfig.KeyStatementName: testStatementName}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1.SqlV1Statement{
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
		},
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     flinkgatewayv1.PtrString(selectFromStatement),
		},
	}
	returnedError := fmt.Errorf("test error")

	client.EXPECT().CreateStatement(SqlV1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, returnedError)

	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: testStatusDetailMessage,
	}

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestProcessStatementFailsOnErrorOnPrem() {
	client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
	store := StoreOnPrem{
		Properties: NewUserPropertiesWithDefaults(
			map[string]string{flinkconfig.KeyStatementName: testStatementName}, map[string]string{},
		),
		client: client,
		appOptions: &types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		},
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := cmfsdk.Statement{
		Status: &cmfsdk.StatementStatus{
			Detail: cmfsdk.PtrString(testStatusDetailMessage),
		},
		Spec: cmfsdk.StatementSpec{
			Properties: &nonLocalProperties,
			Statement:  selectFromStatement,
		},
	}
	returnedError := fmt.Errorf("test error")

	client.EXPECT().CmfApiContext().Return(context.Background())
	client.EXPECT().CreateStatement(context.Background(), "envId", CmfStatementMatcher{statementObj}).Return(statementObj, returnedError)

	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: testStatusDetailMessage,
	}

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestProcessStatementUsesUserProvidedStatementName() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
	}
	serviceAccountId := "sa-123"
	store := Store{
		Properties:       NewUserPropertiesWithDefaults(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId}, map[string]string{flinkconfig.KeyStatementName: testStatementName}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementObj := flinkgatewayv1.SqlV1Statement{
		Name: flinkgatewayv1.PtrString(testStatementName),
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Phase:  "PENDING",
			Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
		},
		Spec: &flinkgatewayv1.SqlV1StatementSpec{
			Properties:    &map[string]string{}, // only sql properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     flinkgatewayv1.PtrString(selectFromStatement),
		},
	}

	client.EXPECT().CreateStatement(SqlV1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
	// statement name should be cleared after submission
	require.False(s.T(), store.Properties.HasKey(flinkconfig.KeyStatementName))
}

func (s *StoreTestSuite) TestProcessStatementUsesUserProvidedStatementNameOnPrem() {
	client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
	store := StoreOnPrem{
		Properties: NewUserPropertiesWithDefaults(map[string]string{}, map[string]string{flinkconfig.KeyStatementName: testStatementName}),
		client:     client,
		appOptions: &types.ApplicationOptions{
			EnvironmentId:   "envId",
			EnvironmentName: "envName",
			Database:        "database",
		},
		tokenRefreshFunc: tokenRefreshFunc,
	}

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := cmfsdk.Statement{
		Status: &cmfsdk.StatementStatus{
			Phase:  "PENDING",
			Detail: cmfsdk.PtrString(testStatusDetailMessage),
		},
		Spec: cmfsdk.StatementSpec{
			Properties: &nonLocalProperties,
			Statement:  selectFromStatement,
		},
	}

	client.EXPECT().CmfApiContext().Return(context.Background())
	client.EXPECT().CreateStatement(context.Background(), "envId", CmfStatementMatcher{statementObj}).Return(statementObj, nil)

	var processedStatement *types.ProcessedStatement
	var err *types.StatementError
	stdout := test.RunAndCaptureSTDOUT(s.T(), func() {
		processedStatement, err = store.ProcessStatement(selectFromStatement)
	})

	cupaloy.SnapshotT(s.T(), stdout)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatementOnPrem(statementObj), processedStatement)
	// statement name should be cleared after submission
	require.False(s.T(), store.Properties.HasKey(flinkconfig.KeyStatementName))
}

func (s *StoreTestSuite) TestWaitPendingStatement() {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		appOptions := &types.ApplicationOptions{
			OrganizationId: "orgId",
			EnvironmentId:  "envId",
		}
		store := Store{
			Properties:       NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:           client,
			appOptions:       appOptions,
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), err)
		require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "COMPLETED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), err)
		require.Equal(s.T(), types.NewProcessedStatementOnPrem(statementObj), processedStatement)
	}
}

func (s *StoreTestSuite) TestWaitPendingStatementNoWaitForCompletedStatement() {
	statement := types.ProcessedStatement{
		Status: types.PHASE("COMPLETED"),
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
		}

		processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &statement, processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
		}

		processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &statement, processedStatement)
	}
}

func (s *StoreTestSuite) TestWaitPendingStatementNoWaitForRunningStatement() {
	statement := types.ProcessedStatement{Status: types.PHASE("RUNNING")}
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
		}

		processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &statement, processedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
		}

		processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
		require.Nil(s.T(), err)
		require.Equal(s.T(), &statement, processedStatement)
	}
}

func (s *StoreTestSuite) TestWaitPendingStatementFailsOnWaitError() {
	returnedErr := fmt.Errorf("test error")
	expectedError := &types.StatementError{
		Message:        returnedErr.Error(),
		FailureMessage: testStatusDetailMessage,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, returnedErr)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, returnedErr)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
}

func (s *StoreTestSuite) TestWaitPendingStatementFailsOnNonCompletedOrRunningStatementPhase() {
	expectedError := &types.StatementError{
		Message:        "can't fetch results. Statement phase is: FAILED",
		FailureMessage: testStatusDetailMessage,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "FAILED",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}

		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Status: &cmfsdk.StatementStatus{
				Phase:  "FAILED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
}

func (s *StoreTestSuite) TestWaitPendingStatementFetchesExceptionOnFailedStatementWithEmptyStatusDetail() {
	exception1 := "Exception 1"
	exception2 := "Exception 2"
	expectedError := &types.StatementError{
		Message:        "can't fetch results. Statement phase is: FAILED",
		FailureMessage: exception1,
	}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		appOptions := &types.ApplicationOptions{
			OrganizationId: "orgId",
			EnvironmentId:  "envId",
		}
		store := Store{
			Properties:       NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:           client,
			appOptions:       appOptions,
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase: "FAILED",
			},
		}
		exceptionsResponse := []flinkgatewayv1.SqlV1StatementException{
			{Message: &exception1},
			{Message: &exception2},
		}

		client.EXPECT().GetStatement("envId", testStatementName, "orgId").Return(statementObj, nil)
		client.EXPECT().GetExceptions("envId", testStatementName, "orgId").Return(exceptionsResponse, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase: "FAILED",
			},
		}

		exceptionsResponse := cmfsdk.StatementExceptionList{
			Data: []cmfsdk.StatementException{
				{Message: exception1},
				{Message: exception2},
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background()).Times(2)
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil)
		client.EXPECT().ListStatementExceptions(context.Background(), "envId", testStatementName).Return(exceptionsResponse, nil)

		processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
			StatementName: testStatementName,
			Status:        types.PENDING,
		})
		require.Nil(s.T(), processedStatement)
		require.Equal(s.T(), expectedError, err)
	}
}

func (s *StoreTestSuite) TestGetStatusDetail() {
	exception1 := "Exception 1"
	exception2 := "Exception 2"

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase: "FAILED",
			},
		}

		exceptionsResponse := []flinkgatewayv1.SqlV1StatementException{
			{Message: &exception1},
			{Message: &exception2},
		}

		client.EXPECT().GetExceptions("envId", testStatementName, "orgId").Return(exceptionsResponse, nil).Times(2)

		require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "FAILING"
		require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase: "FAILED",
			},
		}

		exceptionsResponse := cmfsdk.StatementExceptionList{
			Data: []cmfsdk.StatementException{
				{Message: exception1},
				{Message: exception2},
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background()).Times(2)
		client.EXPECT().ListStatementExceptions(context.Background(), "envId", testStatementName).Return(exceptionsResponse, nil).Times(2)

		require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "FAILING"
		require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
	}
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsWhenStatusNoFailedOrFailing() {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "PENDING",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}

		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "COMPLETED"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "RUNNING"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "DELETED"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase:  "PENDING",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "COMPLETED"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "RUNNING"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
		statementObj.Status.Phase = "DELETED"
		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
	}
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsWhenStatusDetailFilled() {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		store := Store{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "FAILED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase:  "FAILED",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
	}
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsEmptyWhenNoExceptionsAvailable() {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
		appOptions := &types.ApplicationOptions{
			OrganizationId: "orgId",
			EnvironmentId:  "envId",
		}
		store := Store{
			Properties:       NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:           client,
			appOptions:       appOptions,
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name:   flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{Phase: "FAILED"},
		}
		exceptionsResponse := []flinkgatewayv1.SqlV1StatementException{}

		client.EXPECT().GetExceptions("envId", testStatementName, "orgId").Return(exceptionsResponse, nil)

		require.Equal(s.T(), "", store.getStatusDetail(statementObj))
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(s.T()))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{Phase: "FAILED"},
		}
		exceptionsResponse := cmfsdk.StatementExceptionList{}

		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().ListStatementExceptions(context.Background(), "envId", testStatementName).Return(exceptionsResponse, nil)

		require.Equal(s.T(), "", store.getStatusDetail(statementObj))
	}
}

func (s *StoreTestSuite) TestIsSelectStatement() {
	tests := []struct {
		name              string
		statement         flinkgatewayv1.SqlV1Statement
		isSelectStatement bool
	}{
		{
			name:              "select lowercase",
			statement:         createStatementWithSqlKind("select"),
			isSelectStatement: true,
		},
		{
			name:              "select uppercase",
			statement:         createStatementWithSqlKind("SELECT"),
			isSelectStatement: true,
		},
		{
			name:              "select random case",
			statement:         createStatementWithSqlKind("SeLeCt"),
			isSelectStatement: true,
		},
		{
			name:              "leading and trailing white space",
			statement:         createStatementWithSqlKind("   select   "),
			isSelectStatement: false,
		},
		{
			name:              "missing last char",
			statement:         createStatementWithSqlKind("selec"),
			isSelectStatement: false,
		},
		{
			name: "select random case without trait",
			statement: flinkgatewayv1.SqlV1Statement{
				Spec: &flinkgatewayv1.SqlV1StatementSpec{
					Statement: flinkgatewayv1.PtrString("SeLeCt"),
				}},
			isSelectStatement: true,
		},
	}

	for _, testCase := range tests {
		s.T().Run(testCase.name, func(t *testing.T) {
			processedStatement := types.NewProcessedStatement(testCase.statement)
			require.Equal(t, testCase.isSelectStatement, processedStatement.IsSelectStatement())
		})
	}
}

func createStatementWithSqlKind(sqlKind string) flinkgatewayv1.SqlV1Statement {
	return flinkgatewayv1.SqlV1Statement{
		Status: &flinkgatewayv1.SqlV1StatementStatus{
			Traits: &flinkgatewayv1.SqlV1StatementTraits{
				SqlKind: flinkgatewayv1.PtrString(sqlKind),
			},
		},
	}
}

func TestWaitForTerminalStateDoesNotStartWhenAlreadyInTerminalState(t *testing.T) {
	processedStatement := types.ProcessedStatement{Status: types.COMPLETED}

	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		returnedStatement, err := s.WaitForTerminalStatementState(context.Background(), processedStatement)

		assert.Nil(t, err)
		assert.Equal(t, &processedStatement, returnedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		store := StoreOnPrem{
			Properties: NewUserPropertiesWithDefaults(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
			client:     client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		returnedStatement, err := store.WaitForTerminalStatementState(context.Background(), processedStatement)

		assert.Nil(t, err)
		assert.Equal(t, &processedStatement, returnedStatement)
	}
}

func TestWaitForTerminalStateStopsWhenTerminalState(t *testing.T) {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "COMPLETED",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", statementObj.GetName(), "orgId").Return(statementObj, nil)
		processedStatement := types.NewProcessedStatement(statementObj)
		processedStatement.Status = types.RUNNING

		returnedStatement, err := s.WaitForTerminalStatementState(context.Background(), *processedStatement)

		assert.Nil(t, err)
		assert.Equal(t, types.NewProcessedStatement(statementObj), returnedStatement)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		store := StoreOnPrem{
			client: client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase: "COMPLETED",
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil)
		processedStatement := types.NewProcessedStatementOnPrem(statementObj)
		processedStatement.Status = types.RUNNING

		returnedStatement, err := store.WaitForTerminalStatementState(context.Background(), *processedStatement)

		assert.Nil(t, err)
		assert.Equal(t, types.NewProcessedStatementOnPrem(statementObj), returnedStatement)
	}
}

func TestWaitForTerminalStateStopsWhenUserDetaches(t *testing.T) {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}
		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "RUNNING",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", statementObj.GetName(), "orgId").Return(statementObj, nil).AnyTimes()
		ctx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			time.Sleep(2 * time.Second)
			cancelFunc()
		}()

		returnedStatement, err := s.WaitForTerminalStatementState(ctx, *types.NewProcessedStatement(statementObj))

		assert.Nil(t, err)
		assert.Equal(t, types.NewProcessedStatement(statementObj), returnedStatement)
		assert.NotNil(t, ctx.Err())
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		store := StoreOnPrem{
			client: client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}

		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase:  "RUNNING",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}

		client.EXPECT().CmfApiContext().Return(context.Background()).AnyTimes()
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, nil).AnyTimes()
		ctx, cancelFunc := context.WithCancel(context.Background())
		go func() {
			time.Sleep(2 * time.Second)
			cancelFunc()
		}()

		returnedStatement, err := store.WaitForTerminalStatementState(ctx, *types.NewProcessedStatementOnPrem(statementObj))

		assert.Nil(t, err)
		assert.Equal(t, types.NewProcessedStatementOnPrem(statementObj), returnedStatement)
		assert.NotNil(t, ctx.Err())
	}
}

func TestWaitForTerminalStateStopsOnError(t *testing.T) {
	{ // Cloud store
		client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
		s := &Store{
			client: client,
			appOptions: &types.ApplicationOptions{
				OrganizationId: "orgId",
				EnvironmentId:  "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}
		statementObj := flinkgatewayv1.SqlV1Statement{
			Name: flinkgatewayv1.PtrString(testStatementName),
			Status: &flinkgatewayv1.SqlV1StatementStatus{
				Phase:  "RUNNING",
				Detail: flinkgatewayv1.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().GetStatement("envId", statementObj.GetName(), "orgId").Return(statementObj, fmt.Errorf("error"))

		_, err := s.WaitForTerminalStatementState(context.Background(), *types.NewProcessedStatement(statementObj))

		assert.NotNil(t, err)
	}
	{ // On-prem store
		client := mock.NewMockCmfClientInterface(gomock.NewController(t))
		store := StoreOnPrem{
			client: client,
			appOptions: &types.ApplicationOptions{
				EnvironmentId: "envId",
			},
			tokenRefreshFunc: tokenRefreshFunc,
		}
		statementObj := cmfsdk.Statement{
			Metadata: cmfsdk.StatementMetadata{
				Name: testStatementName,
			},
			Status: &cmfsdk.StatementStatus{
				Phase:  "RUNNING",
				Detail: cmfsdk.PtrString(testStatusDetailMessage),
			},
		}
		client.EXPECT().CmfApiContext().Return(context.Background())
		client.EXPECT().GetStatement(context.Background(), "envId", testStatementName).Return(statementObj, fmt.Errorf("error"))

		_, err := store.WaitForTerminalStatementState(context.Background(), *types.NewProcessedStatementOnPrem(statementObj))

		assert.NotNil(t, err)
	}
}

type SqlV1StatementMatcher struct {
	Expected flinkgatewayv1.SqlV1Statement
}

func (p SqlV1StatementMatcher) Matches(x interface{}) bool {
	actual, ok := x.(flinkgatewayv1.SqlV1Statement)
	if !ok {
		return false
	}
	statementMatches := *actual.Spec.ComputePoolId == *p.Expected.Spec.ComputePoolId &&
		reflect.DeepEqual(actual.Spec.Properties, p.Expected.Spec.Properties) &&
		*actual.Spec.Statement == *p.Expected.Spec.Statement
	if p.Expected.Name == nil {
		return statementMatches
	}
	return statementMatches && *actual.Name == *p.Expected.Name
}

func (p SqlV1StatementMatcher) String() string {
	return fmt.Sprintf("%v", p.Expected)
}

type CmfStatementMatcher struct {
	Expected cmfsdk.Statement
}

func (p CmfStatementMatcher) Matches(x interface{}) bool {
	actual, ok := x.(cmfsdk.Statement)
	if !ok {
		return false
	}
	statementMatches := actual.Spec.ComputePoolName == p.Expected.Spec.ComputePoolName &&
		reflect.DeepEqual(actual.Spec.Properties, p.Expected.Spec.Properties) &&
		actual.Spec.Statement == p.Expected.Spec.Statement
	if p.Expected.Metadata.Name == "" {
		return statementMatches
	}
	return statementMatches && actual.Metadata.Name == p.Expected.Metadata.Name
}

func (p CmfStatementMatcher) String() string {
	return fmt.Sprintf("%v", p.Expected)
}
