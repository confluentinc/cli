package store

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"go.uber.org/mock/gomock"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	flinkgatewayv1beta1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1beta1"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors/flink"
	flinkconfig "github.com/confluentinc/cli/v3/pkg/flink/config"
	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	testserver "github.com/confluentinc/cli/v3/test/test-server"
)

type StoreTestSuite struct {
	suite.Suite
}

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
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	mockAppController := mock.NewMockApplicationControllerInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	s := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc).(*Store)

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
}

func TestWaitForPendingStatement3(t *testing.T) {
	statementName := "statementName"

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

	// Test case 1: Statement is not pending
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "COMPLETED",
			Detail: flinkgatewayv1beta1.PtrString("Test status detail message"),
		},
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil)

	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, time.Duration(10))
	assert.Nil(t, err)
	assert.NotNil(t, processedStatement)
	assert.Equal(t, types.NewProcessedStatement(statementObj), processedStatement)
}

func TestWaitForPendingTimesout(t *testing.T) {
	statementName := "statementName"
	timeout := time.Duration(10) * time.Millisecond

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

	statusDetailMessage := "test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
	}
	expectedError := &types.StatementError{
		Message: fmt.Sprintf("statement is still pending after %f seconds. If you want to increase the timeout for the client, you can run \"SET '%s'='10000';\" to adjust the maximum timeout in milliseconds.",
			timeout.Seconds(), flinkconfig.KeyResultsTimeout),
		FailureMessage: fmt.Sprintf("captured retryable errors: %s", statusDetailMessage),
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil).AnyTimes()
	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, timeout)

	assert.Equal(t, expectedError, err)
	assert.Nil(t, processedStatement)
}

func TestWaitForPendingHitsErrorRetryLimit(t *testing.T) {
	statementName := "statementName"
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

	statusDetailMessage := "test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
	}
	expectedError := &types.StatementError{
		Message:        "the server can't process this statement right now, exiting after 6 retries",
		FailureMessage: fmt.Sprintf("captured retryable errors: %s", strings.Repeat(statusDetailMessage+"; ", 5)+statusDetailMessage),
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil).AnyTimes()
	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, timeout)

	assert.Equal(t, expectedError, err)
	assert.Nil(t, processedStatement)
}

func TestWaitForPendingEventuallyCompletes(t *testing.T) {
	statementName := "statementName"

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

	transientStatusDetailMessage := "Transient status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &transientStatusDetailMessage,
		},
	}

	finalStatusDetailMessage := "Final status detail message"
	statementObjCompleted := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "COMPLETED",
			Detail: &finalStatusDetailMessage,
		},
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil).Times(3)
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObjCompleted, nil)

	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, time.Duration(10)*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, processedStatement)
	assert.Equal(t, types.NewProcessedStatement(statementObjCompleted), processedStatement)
}

func TestWaitForPendingStatementErrors(t *testing.T) {
	statementName := "statementName"
	waitTime := time.Millisecond * 1
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
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "COMPLETED",
			Detail: &statusDetailMessage,
		},
	}

	returnedError := flink.NewError("couldn't get statement", "", http.StatusInternalServerError)
	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: statusDetailMessage,
		StatusCode:     http.StatusInternalServerError,
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, returnedError)
	_, err := s.waitForPendingStatement(context.Background(), statementName, waitTime)
	assert.Equal(t, expectedError, err)
}

func TestCancelPendingStatement(t *testing.T) {
	statementName := "statementName"
	waitTime := time.Second * 1
	ctx, cancelFunc := context.WithCancel(context.Background())

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

	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase: "PENDING",
		},
	}

	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	expectedErr := &types.StatementError{Message: "result retrieval aborted. Statement will be deleted", StatusCode: http.StatusInternalServerError}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil).AnyTimes()
	client.EXPECT().DeleteStatement("envId", statementName, "orgId").Return(nil).AnyTimes()
	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return([]flinkgatewayv1beta1.SqlV1beta1StatementException{}, flinkError).AnyTimes()

	// Schedule routine to cancel context
	go func() {
		time.Sleep(time.Millisecond * 20)
		cancelFunc()
	}()

	res, err := s.waitForPendingStatement(ctx, statementName, waitTime)

	assert.Nil(t, res)
	assert.EqualError(t, err, expectedErr.Error())
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
		Message: `"=" should only appear once`,
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement("SET key=value")
	assert.Equal(s.T(), &types.StatementError{
		Message: "key and value must be enclosed by single quotes (')",
		Usage:   []string{"SET 'key'='value'"},
	}, err)

	_, _, err = parseSetStatement(`set ''key'''=''value''`)
	assert.Equal(s.T(), &types.StatementError{
		Message:    "key contains unescaped single quotes (')",
		Usage:      []string{"SET 'key'='value'"},
		Suggestion: `please escape all single quotes with another single quote "''key''"`,
	}, err)

	_, _, err = parseSetStatement(`set 'key'=''value''`)
	assert.Equal(s.T(), &types.StatementError{
		Message:    "value contains unescaped single quotes (')",
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
	statementName := "TEST_STATEMENT"
	statementObj := flinkgatewayv1beta1.NewSqlV1beta1StatementWithDefaults()
	spec := flinkgatewayv1beta1.NewSqlV1beta1StatementSpecWithDefaults()
	statementObj.SetName(statementName)
	statementObj.SetSpec(*spec)

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(*statementObj, nil)

	statementUpdated := flinkgatewayv1beta1.NewSqlV1beta1StatementWithDefaults()
	specUpdated := flinkgatewayv1beta1.NewSqlV1beta1StatementSpecWithDefaults()
	statementUpdated.SetName(statementName)
	specUpdated.SetStopped(true)
	statementUpdated.SetSpec(*specUpdated)

	client.EXPECT().UpdateStatement("envId", statementName, "orgId", *statementUpdated).Return(nil)

	wasStatementDeleted := store.StopStatement(statementName)
	require.True(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestStopStatementFailsOnGetError() {
	ctrl := gomock.NewController(s.T())
	statementName := "TEST_STATEMENT"

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(flinkgatewayv1beta1.SqlV1beta1Statement{}, flinkError)

	wasStatementDeleted := store.StopStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestStopStatementFailsOnNilSpecError() {
	ctrl := gomock.NewController(s.T())
	statementName := "TEST_STATEMENT"
	statementObj := flinkgatewayv1beta1.NewSqlV1beta1StatementWithDefaults()
	statementObj.SetName(statementName)

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(flinkgatewayv1beta1.SqlV1beta1Statement{}, flinkError)

	wasStatementDeleted := store.StopStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestStopStatementFailsOnUpdateError() {
	ctrl := gomock.NewController(s.T())
	statementName := "TEST_STATEMENT"
	statementObj := flinkgatewayv1beta1.NewSqlV1beta1StatementWithDefaults()
	spec := flinkgatewayv1beta1.NewSqlV1beta1StatementSpecWithDefaults()
	statementObj.SetName(statementName)
	statementObj.SetSpec(*spec)

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(*statementObj, nil)
	statementObj.Spec.SetStopped(true)
	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	client.EXPECT().UpdateStatement("envId", statementName, "orgId", *statementObj).Return(flinkError)

	wasStatementDeleted := store.StopStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestDeleteStatement() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement("envId", statementName, "orgId").Return(nil)

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.True(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestDeleteStatementFailsOnError() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statementName := "TEST_STATEMENT"

	flinkError := flink.NewError("error", "", http.StatusInternalServerError)
	client.EXPECT().DeleteStatement("envId", statementName, "orgId").Return(flinkError)
	wasStatementDeleted := store.DeleteStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWithCompletedStatement() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.COMPLETED,
	}
	statementResultObj := flinkgatewayv1beta1.SqlV1beta1StatementResult{
		Metadata: flinkgatewayv1beta1.ResultListMeta{},
		Results:  &flinkgatewayv1beta1.SqlV1beta1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsWithRunningStatement() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := flinkgatewayv1beta1.SqlV1beta1StatementResult{
		Metadata: flinkgatewayv1beta1.ResultListMeta{},
		Results:  &flinkgatewayv1beta1.SqlV1beta1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenPageTokenExists() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	nextPage := "https://devel.cpdev.cloud/some/results?page_token=eyJWZX"
	statementResultObj := flinkgatewayv1beta1.SqlV1beta1StatementResult{
		Metadata: flinkgatewayv1beta1.ResultListMeta{Next: &nextPage},
		Results:  &flinkgatewayv1beta1.SqlV1beta1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenResultsExist() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrganizationId:  "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := flinkgatewayv1beta1.SqlV1beta1StatementResult{
		Metadata: flinkgatewayv1beta1.ResultListMeta{},
		Results:  &flinkgatewayv1beta1.SqlV1beta1StatementResultResults{Data: &[]any{map[string]any{"op": 0}}},
	}
	client.EXPECT().GetStatementResults("envId", statement.StatementName, "orgId", statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
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
		store := Store{Properties: NewUserProperties(tc.properties, map[string]string{})}
		result := store.getTimeout()
		require.Equal(t, tc.expected, result, tc.name)
	}
}

func (s *StoreTestSuite) TestProcessStatementWithServiceAccount() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
	}
	serviceAccountId := "sa-123"
	store := Store{
		Properties:       NewUserProperties(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId, "TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statement := "SELECT * FROM table"
	statusDetailMessage := "Test status detail message"

	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
		Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     &statement,
		},
	}

	client.EXPECT().CreateStatement(SqlV1beta1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	processedStatement, err := store.ProcessStatement(statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
}

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
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statement := "SELECT * FROM table"
	statusDetailMessage := "Test status detail message"
	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
		Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     &statement,
		},
	}

	client.EXPECT().CreateStatement(SqlV1beta1StatementMatcher{statementObj}, user, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	processedStatement, err := store.ProcessStatement(statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
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
		Properties:       NewUserProperties(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId, "TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statement := "SELECT * FROM table"
	statusDetailMessage := "test status detail message"
	nonLocalProperties := store.Properties.GetNonLocalProperties()
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Detail: &statusDetailMessage,
		},
		Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
			Properties:    &nonLocalProperties, // only non-local properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     &statement,
		},
	}
	returnedError := fmt.Errorf("test error")

	client.EXPECT().CreateStatement(SqlV1beta1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, returnedError)

	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: statusDetailMessage,
	}

	processedStatement, err := store.ProcessStatement(statement)
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
	statementName := "test-statement"
	store := Store{
		Properties:       NewUserProperties(map[string]string{flinkconfig.KeyServiceAccount: serviceAccountId}, map[string]string{flinkconfig.KeyStatementName: statementName}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statement := "SELECT * FROM table"
	statusDetailMessage := "Test status detail message"

	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
		Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
			Properties:    &map[string]string{}, // only sql properties are passed to the gateway
			ComputePoolId: &appOptions.ComputePoolId,
			Statement:     &statement,
		},
	}

	client.EXPECT().CreateStatement(SqlV1beta1StatementMatcher{statementObj}, serviceAccountId, appOptions.EnvironmentId, appOptions.OrganizationId).
		Return(statementObj, nil)

	processedStatement, err := store.ProcessStatement(statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
	// statement name should be cleared after submission
	require.False(s.T(), store.Properties.HasKey(flinkconfig.KeyStatementName))
}

func (s *StoreTestSuite) TestWaitPendingStatement() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "COMPLETED",
			Detail: &statusDetailMessage,
		},
	}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil)

	processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
		StatementName: statementName,
		Status:        types.PENDING,
	})
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
}

func (s *StoreTestSuite) TestWaitPendingStatementNoWaitForCompletedStatement() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	store := Store{
		Properties: NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:     client,
	}

	statement := types.ProcessedStatement{
		Status: types.PHASE("COMPLETED"),
	}

	processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), &statement, processedStatement)
}

func (s *StoreTestSuite) TestWaitPendingStatementNoWaitForRunningStatement() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	store := Store{
		Properties: NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:     client,
	}

	statement := types.ProcessedStatement{Status: types.PHASE("RUNNING")}

	processedStatement, err := store.WaitPendingStatement(context.Background(), statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), &statement, processedStatement)
}

func (s *StoreTestSuite) TestWaitPendingStatementFailsOnWaitError() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Detail: &statusDetailMessage,
		},
	}
	returnedErr := fmt.Errorf("test error")
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, returnedErr)
	expectedError := &types.StatementError{
		Message:        returnedErr.Error(),
		FailureMessage: statusDetailMessage,
	}

	processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
		StatementName: statementName,
		Status:        types.PENDING,
	})
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestWaitPendingStatementFailsOnNonCompletedOrRunningStatementPhase() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "FAILED",
			Detail: &statusDetailMessage,
		},
	}
	expectedError := &types.StatementError{
		Message:        fmt.Sprintf("can't fetch results. Statement phase is: %s", statementObj.Status.Phase),
		FailureMessage: statusDetailMessage,
	}

	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil)

	processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
		StatementName: statementName,
		Status:        types.PENDING,
	})
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestWaitPendingStatementFetchesExceptionOnFailedStatementWithEmptyStatusDetail() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase: "FAILED",
		},
	}
	exception1 := "Exception 1"
	exception2 := "Exception 2"
	exceptionsResponse := []flinkgatewayv1beta1.SqlV1beta1StatementException{
		{Stacktrace: &exception1},
		{Stacktrace: &exception2},
	}
	expectedError := &types.StatementError{
		Message:        fmt.Sprintf("can't fetch results. Statement phase is: %s", statementObj.Status.Phase),
		FailureMessage: exception1,
	}

	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil)
	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return(exceptionsResponse, nil)

	processedStatement, err := store.WaitPendingStatement(context.Background(), types.ProcessedStatement{
		StatementName: statementName,
		Status:        types.PENDING,
	})
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestGetStatusDetail() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase: "FAILED",
		},
	}
	exception1 := "Exception 1"
	exception2 := "Exception 2"
	exceptionsResponse := []flinkgatewayv1beta1.SqlV1beta1StatementException{
		{Stacktrace: &exception1},
		{Stacktrace: &exception2},
	}

	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return(exceptionsResponse, nil).Times(2)

	require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
	statementObj.Status.Phase = "FAILING"
	require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsWhenStatusNoFailedOrFailing() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	testStatusDetailMessage := "Test Status Detail Message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString("Test Statement"),
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "PENDING",
			Detail: &testStatusDetailMessage,
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

func (s *StoreTestSuite) TestGetStatusDetailReturnsWhenStatusDetailFilled() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	testStatusDetailMessage := "Test Status Detail Message"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString("Test Statement"),
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "FAILED",
			Detail: &testStatusDetailMessage,
		},
	}

	require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsEmptyWhenNoExceptionsAvailable() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrganizationId: "orgId",
		EnvironmentId:  "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}, map[string]string{}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name:   &statementName,
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{Phase: "FAILED"},
	}
	exceptionsResponse := []flinkgatewayv1beta1.SqlV1beta1StatementException{}

	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return(exceptionsResponse, nil)

	require.Equal(s.T(), "", store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestNewProcessedStatementSetsIsSelectStatement() {
	tests := []struct {
		name              string
		statement         flinkgatewayv1beta1.SqlV1beta1Statement
		isSelectStatement bool
	}{
		{
			name: "select lowercase",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("select * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "select uppercase",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("SELECT * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "select random case",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("SeLeCt * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "leading white space",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("   select * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "missing last char",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("selec * FROM table"),
				},
			},
			isSelectStatement: false,
		},
		{
			name: "missing last char",
			statement: flinkgatewayv1beta1.SqlV1beta1Statement{
				Spec: &flinkgatewayv1beta1.SqlV1beta1StatementSpec{
					Statement: flinkgatewayv1beta1.PtrString("insert into table values (1, 2)"),
				},
			},
			isSelectStatement: false,
		},
	}

	for _, testCase := range tests {
		s.T().Run(testCase.name, func(t *testing.T) {
			processedStatement := types.NewProcessedStatement(testCase.statement)
			require.Equal(t, testCase.isSelectStatement, processedStatement.IsSelectStatement)
		})
	}
}

func TestWaitForTerminalStateDoesNotStartWhenAlreadyInTerminalState(t *testing.T) {
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
	processedStatement := types.ProcessedStatement{Status: types.COMPLETED}

	returnedStatement, err := s.WaitForTerminalStatementState(context.Background(), processedStatement)

	assert.Nil(t, err)
	assert.Equal(t, &processedStatement, returnedStatement)
}

func TestWaitForTerminalStateStopsWhenTerminalState(t *testing.T) {
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

	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString("statement-name"),
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "COMPLETED",
			Detail: flinkgatewayv1beta1.PtrString("Test status detail message"),
		},
	}
	client.EXPECT().GetStatement("envId", statementObj.GetName(), "orgId").Return(statementObj, nil)
	processedStatement := types.NewProcessedStatement(statementObj)
	processedStatement.Status = types.RUNNING

	returnedStatement, err := s.WaitForTerminalStatementState(context.Background(), *processedStatement)

	assert.Nil(t, err)
	assert.Equal(t, types.NewProcessedStatement(statementObj), returnedStatement)
}

func TestWaitForTerminalStateStopsWhenUserDetaches(t *testing.T) {
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
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString("statement-name"),
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "RUNNING",
			Detail: flinkgatewayv1beta1.PtrString("Test status detail message"),
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

func TestWaitForTerminalStateStopsOnError(t *testing.T) {
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
	statementObj := flinkgatewayv1beta1.SqlV1beta1Statement{
		Name: flinkgatewayv1beta1.PtrString("statement-name"),
		Status: &flinkgatewayv1beta1.SqlV1beta1StatementStatus{
			Phase:  "RUNNING",
			Detail: flinkgatewayv1beta1.PtrString("Test status detail message"),
		},
	}
	client.EXPECT().GetStatement("envId", statementObj.GetName(), "orgId").Return(statementObj, fmt.Errorf("error"))

	_, err := s.WaitForTerminalStatementState(context.Background(), *types.NewProcessedStatement(statementObj))

	assert.NotNil(t, err)
}

type SqlV1beta1StatementMatcher struct {
	Expected flinkgatewayv1beta1.SqlV1beta1Statement
}

func (p SqlV1beta1StatementMatcher) Matches(x interface{}) bool {
	actual, ok := x.(flinkgatewayv1beta1.SqlV1beta1Statement)
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

func (p SqlV1beta1StatementMatcher) String() string {
	return fmt.Sprintf("%v", p.Expected)
}
