package store

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	flinkgatewayv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/test/mock"
	"github.com/confluentinc/cli/internal/pkg/flink/types"
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

func TestStoreProcessLocalStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	mockAppController := mock.NewMockApplicationControllerInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrgResourceId:   "orgId",
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

	result, err = s.ProcessLocalStatement("USE my_database;")
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	// Test case 1: Statement is not pending
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "COMPLETED",
			Detail: flinkgatewayv1alpha1.PtrString("Test status detail message"),
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statusDetailMessage := "test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
	}
	expectedError := &types.StatementError{
		Message:        fmt.Sprintf("statement is still pending after %f seconds. If you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.", timeout.Seconds()),
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statusDetailMessage := "test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	transientStatusDetailMessage := "Transient status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "PENDING",
			Detail: &transientStatusDetailMessage,
		},
	}

	finalStatusDetailMessage := "Final status detail message"
	statementObjCompleted := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "COMPLETED",
			Detail: &statusDetailMessage,
		},
	}

	returnedError := errors.New("couldn't get statement")
	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: statusDetailMessage,
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	s := &Store{
		client:           client,
		appOptions:       &appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}

	expectedErr := &types.StatementError{Message: "result retrieval aborted. Statement will be deleted"}
	client.EXPECT().GetStatement("envId", statementName, "orgId").Return(statementObj, nil).AnyTimes()

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
	assert.True(s.T(), true, statementStartsWithOp("SET", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key=value", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET   ", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    set   ", config.ConfigOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", config.ConfigOpSet))

	assert.False(s.T(), false, statementStartsWithOp("SETting", config.ConfigOpSet))
	assert.False(s.T(), false, statementStartsWithOp("", config.ConfigOpSet))
	assert.False(s.T(), false, statementStartsWithOp("should be false", config.ConfigOpSet))
	assert.False(s.T(), false, statementStartsWithOp("USE", config.ConfigOpSet))
	assert.False(s.T(), false, statementStartsWithOp("SETTING", config.ConfigOpSet))
}

func (s *StoreTestSuite) TestIsUseStatement() {
	assert.True(s.T(), statementStartsWithOp("USE", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("USE catalog", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("use CATALOG cat", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("USE   ", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("use   ", config.ConfigOpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", config.ConfigOpUse))

	assert.False(s.T(), statementStartsWithOp("SET", config.ConfigOpUse))
	assert.False(s.T(), statementStartsWithOp("USES", config.ConfigOpUse))
	assert.False(s.T(), statementStartsWithOp("", config.ConfigOpUse))
	assert.False(s.T(), statementStartsWithOp("should be false", config.ConfigOpUse))
}

func (s *StoreTestSuite) TestIsResetStatement() {
	assert.True(s.T(), true, statementStartsWithOp("RESET", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET   ", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("reset   ", config.ConfigOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", config.ConfigOpReset))

	assert.False(s.T(), false, statementStartsWithOp("RESETting", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("USE", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("RESETTING", config.ConfigOpReset))
}

func (s *StoreTestSuite) TestIsExitStatement() {
	assert.True(s.T(), true, statementStartsWithOp("EXIT", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("EXIT ;", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit   ;", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exiT   ", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("Exit   ", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("eXit   ", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit", config.ConfigOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit ", config.ConfigOpExit))

	assert.False(s.T(), false, statementStartsWithOp("exits", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("exitt;", config.ConfigOpReset))
	assert.False(s.T(), false, statementStartsWithOp("exi", config.ConfigOpReset))
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
	key, value, _ := parseUseStatement("USE CATALOG c;")
	assert.Equal(s.T(), config.ConfigKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value, _ = parseUseStatement("use   catalog   \nc   ")
	assert.Equal(s.T(), config.ConfigKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value, _ = parseUseStatement("use   catalog     ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("catalog   c")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("use     db   ")
	assert.Equal(s.T(), config.ConfigKeyDatabase, key)
	assert.Equal(s.T(), "db", value)

	key, value, _ = parseUseStatement("dAtaBaSe  db   ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("use     \ndatabase_name   ")
	assert.Equal(s.T(), config.ConfigKeyDatabase, key)
	assert.Equal(s.T(), "database_name", value)
}

func (s *StoreTestSuite) TestParseUseStatementError() {
	_, _, err := parseUseStatement("USE CATALOG ;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: missing catalog name\nUsage: \"USE CATALOG my_catalog\"", err.Error())

	_, _, err = parseUseStatement("USE;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: missing database/catalog name\nUsage: \"USE CATALOG my_catalog\" or \"USE my_database\"", err.Error())

	_, _, err = parseUseStatement("USE CATALOG DATABASE DB2;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax for USE\nUsage: \"USE CATALOG my_catalog\" or \"USE my_database\"", err.Error())
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

func (s *StoreTestSuite) TestProccessHttpErrors() {
	// given
	res := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       generateCloserFromObject(flinkgatewayv1alpha1.NewError()),
	}

	// when
	err := processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: unauthorized\nSuggestion: Please run \"confluent login\"", err.Error())

	// given
	title := "invalid syntax"
	detail := "you should provide a table for select"
	statementErr := &flinkgatewayv1alpha1.Error{Title: &title, Detail: &detail}
	res = &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       generateCloserFromObject(statementErr),
	}

	// when
	err = processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: invalid syntax: you should provide a table for select", err.Error())

	// given
	res = &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       generateCloserFromObject(nil),
	}

	// when
	err = processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: received error with code \"500\" from server but could not parse it. This is not expected. Please contact support", err.Error())

	// given
	res = &http.Response{
		StatusCode: http.StatusCreated,
		Body:       generateCloserFromObject(nil),
	}

	// when
	err = processHttpErrors(res, nil)

	// expect
	assert.Nil(s.T(), err)

	// given
	err = errors.New("some error")

	// when
	err = processHttpErrors(nil, err)

	// expect
	assert.Equal(s.T(), "Error: some error", err.Error())
}

func generateCloserFromObject(obj interface{}) io.ReadCloser {
	bts, _ := json.Marshal(obj)
	buf := bytes.NewReader(bts)
	reader := bufio.NewReader(buf)
	return io.NopCloser(reader)
}

func (s *StoreTestSuite) TestDeleteStatement() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId:   "orgId",
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
		OrgResourceId:   "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement("envId", statementName, "orgId").Return(errors.New("test error"))

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWithCompletedStatement() {
	ctrl := gomock.NewController(s.T())

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId:   "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.COMPLETED,
	}
	statementResultObj := flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
		Metadata: flinkgatewayv1alpha1.ResultListMeta{},
		Results:  &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{},
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
		OrgResourceId:   "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
		Metadata: flinkgatewayv1alpha1.ResultListMeta{},
		Results:  &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{},
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
		OrgResourceId:   "orgId",
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
	statementResultObj := flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
		Metadata: flinkgatewayv1alpha1.ResultListMeta{Next: &nextPage},
		Results:  &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{},
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
		OrgResourceId:   "orgId",
		EnvironmentId:   "envId",
		EnvironmentName: "envName",
		Database:        "database",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions, tokenRefreshFunc)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := flinkgatewayv1alpha1.SqlV1alpha1StatementResult{
		Metadata: flinkgatewayv1alpha1.ResultListMeta{},
		Results:  &flinkgatewayv1alpha1.SqlV1alpha1StatementResultResults{Data: &[]any{map[string]any{"op": 0}}},
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
		{0, config.InitialWaitTime},
		{3, config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*0)*time.Millisecond},
		{7, config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*0)*time.Millisecond},
		{10, config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*1)*time.Millisecond},
		{15, config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*1)*time.Millisecond},
		{32, config.InitialWaitTime + time.Duration(config.WaitTimeIncrease*3)*time.Millisecond},
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
				config.ConfigKeyResultsTimeout: "10", // timeout in seconds
			},
			expected: 10 * time.Second,
		},
		{
			name:       "results-timeout property not set",
			properties: map[string]string{},
			expected:   config.DefaultTimeoutDuration,
		},
		{
			name: "invalid results-timeout property",
			properties: map[string]string{
				config.ConfigKeyResultsTimeout: "abc", // invalid duration
			},
			expected: config.DefaultTimeoutDuration,
		},
	}

	// Iterate over test cases and run the function for each input, comparing output to expected value
	for _, tc := range testCases {
		store := Store{Properties: NewUserProperties(tc.properties)}
		result := store.getTimeout()
		require.Equal(t, tc.expected, result, tc.name)
	}
}

func (s *StoreTestSuite) TestProcessStatement() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrgResourceId:  "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
		IdentityPoolId: "identityPoolId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "PENDING",
			Detail: &statusDetailMessage,
		},
	}

	statement := "SELECT * FROM table"
	client.EXPECT().CreateStatement(statement, "computePoolId", "identityPoolId", store.Properties.GetProperties(), "envId", "orgId").
		Return(statementObj, nil)

	processedStatement, err := store.ProcessStatement(statement)
	require.Nil(s.T(), err)
	require.Equal(s.T(), types.NewProcessedStatement(statementObj), processedStatement)
}

func (s *StoreTestSuite) TestProcessStatementFailsOnError() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrgResourceId:  "orgId",
		EnvironmentId:  "envId",
		ComputePoolId:  "computePoolId",
		IdentityPoolId: "identityPoolId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statusDetailMessage := "test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Detail: &statusDetailMessage,
		},
	}
	returnedError := errors.New("test error")

	statement := "SELECT * FROM table"
	client.EXPECT().CreateStatement(statement, "computePoolId", "identityPoolId", store.Properties.GetProperties(), "envId", "orgId").
		Return(statementObj, returnedError)
	expectedError := &types.StatementError{
		Message:        returnedError.Error(),
		FailureMessage: statusDetailMessage,
	}

	processedStatement, err := store.ProcessStatement(statement)
	require.Nil(s.T(), processedStatement)
	require.Equal(s.T(), expectedError, err)
}

func (s *StoreTestSuite) TestWaitPendingStatement() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
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
		Properties: NewUserProperties(map[string]string{"TestProp": "TestVal"}),
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
		Properties: NewUserProperties(map[string]string{"TestProp": "TestVal"}),
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Detail: &statusDetailMessage,
		},
	}
	returnedErr := errors.New("test error")
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statusDetailMessage := "Test status detail message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase: "FAILED",
		},
	}
	exception1 := "Exception 1"
	exception2 := "Exception 2"
	exceptionsResponse := flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList{
		Data: []flinkgatewayv1alpha1.SqlV1alpha1StatementException{
			{Stacktrace: &exception1},
			{Stacktrace: &exception2},
		},
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase: "FAILED",
		},
	}
	exception1 := "Exception 1"
	exception2 := "Exception 2"
	exceptionsResponse := flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList{
		Data: []flinkgatewayv1alpha1.SqlV1alpha1StatementException{
			{Stacktrace: &exception1},
			{Stacktrace: &exception2},
		},
	}

	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return(exceptionsResponse, nil).Times(2)

	require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
	statementObj.Status.Phase = "FAILING"
	require.Equal(s.T(), exception1, store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsWhenStatusNoFailedOrFailing() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	testStatusDetailMessage := "Test Status Detail Message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: flinkgatewayv1alpha1.PtrString("Test Statement"),
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
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
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	testStatusDetailMessage := "Test Status Detail Message"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: flinkgatewayv1alpha1.PtrString("Test Statement"),
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase:  "FAILED",
			Detail: &testStatusDetailMessage,
		},
	}

	require.Equal(s.T(), testStatusDetailMessage, store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestGetStatusDetailReturnsEmptyWhenNoExceptionsAvailable() {
	client := mock.NewMockGatewayClientInterface(gomock.NewController(s.T()))
	appOptions := &types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvironmentId: "envId",
	}
	store := Store{
		Properties:       NewUserProperties(map[string]string{"TestProp": "TestVal"}),
		client:           client,
		appOptions:       appOptions,
		tokenRefreshFunc: tokenRefreshFunc,
	}

	statementName := "Test Statement"
	statementObj := flinkgatewayv1alpha1.SqlV1alpha1Statement{
		Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
			StatementName: &statementName,
		},
		Status: &flinkgatewayv1alpha1.SqlV1alpha1StatementStatus{
			Phase: "FAILED",
		},
	}
	exceptionsResponse := flinkgatewayv1alpha1.SqlV1alpha1StatementExceptionList{
		Data: []flinkgatewayv1alpha1.SqlV1alpha1StatementException{},
	}

	client.EXPECT().GetExceptions("envId", statementName, "orgId").Return(exceptionsResponse, nil)

	require.Equal(s.T(), "", store.getStatusDetail(statementObj))
}

func (s *StoreTestSuite) TestNewProcessedStatementSetsIsSelectStatement() {
	tests := []struct {
		name              string
		statement         flinkgatewayv1alpha1.SqlV1alpha1Statement
		isSelectStatement bool
	}{
		{
			name: "select lowercase",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("select * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "select uppercase",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("SELECT * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "select random case",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("SeLeCt * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "leading white space",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("   select * FROM table"),
				},
			},
			isSelectStatement: true,
		},
		{
			name: "missing last char",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("selec * FROM table"),
				},
			},
			isSelectStatement: false,
		},
		{
			name: "missing last char",
			statement: flinkgatewayv1alpha1.SqlV1alpha1Statement{
				Spec: &flinkgatewayv1alpha1.SqlV1alpha1StatementSpec{
					Statement: flinkgatewayv1alpha1.PtrString("insert into table values (1, 2)"),
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
