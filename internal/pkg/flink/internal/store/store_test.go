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
	"testing"
	"time"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1alpha1"

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

func TestStoreProcessLocalStatement(t *testing.T) {
	// Create a new store
	client := ccloudv2.NewFlinkGatewayClient("url", "userAgent", false, "authToken")
	mockAppController := mock.NewMockApplicationControllerInterface(gomock.NewController(t))
	s := NewStore(client, mockAppController.ExitApplication, nil).(*Store)

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

	mockAppController.EXPECT().ExitApplication().Times(1)
	result, err = s.ProcessLocalStatement("EXIT;")
	assert.Nil(t, err)
	assert.Nil(t, result)
}

func TestWaitForPendingStatement3(t *testing.T) {
	statementName := "statementName"

	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	s := &Store{
		client:     client,
		appOptions: &appOptions,
	}

	// Test case 1: Statement is not pending
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "COMPLETED",
		},
	}
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObj, nil).Times(1)

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
		EnvId:         "envId",
	}
	s := &Store{
		client:     client,
		appOptions: &appOptions,
	}

	// Test case 2: Statement is pending
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObj, nil).AnyTimes()
	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, timeout)

	assert.EqualError(t, err, fmt.Sprintf("Error: Statement is still pending after %f seconds. \n\nIf you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.", timeout.Seconds()))
	assert.Nil(t, processedStatement)
}

func TestWaitForPendingEventuallyCompletes(t *testing.T) {
	statementName := "statementName"

	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	s := &Store{
		client:     client,
		appOptions: &appOptions,
	}

	// Test case 2: Statement is pending
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}

	statementObjCompleted := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "COMPLETED",
		},
	}
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObj, nil).Times(3)
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObjCompleted, nil).Times(1)

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
		EnvId:         "envId",
	}
	s := &Store{
		client:     client,
		appOptions: &appOptions,
	}
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "COMPLETED",
		},
	}

	expectedErr := errors.New("couldn't get statement!")
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObj, expectedErr).Times(1)
	_, err := s.waitForPendingStatement(context.Background(), statementName, waitTime)
	assert.EqualError(t, err, "Error: "+expectedErr.Error())
}

func TestCancelPendingStatement(t *testing.T) {
	statementName := "statementName"
	waitTime := time.Second * 1
	ctx, cancelFunc := context.WithCancel(context.Background())

	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	s := &Store{
		client:     client,
		appOptions: &appOptions,
	}

	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}

	expectedErr := &types.StatementError{Msg: "Result retrieval aborted. Statement will be deleted."}
	client.EXPECT().GetStatement("orgId", "envId", statementName).Return(statementObj, nil).AnyTimes()

	// Schedule routine to cancel context
	go func() {
		time.Sleep(time.Millisecond * 20)
		cancelFunc()
	}()

	res, err := s.waitForPendingStatement(ctx, statementName, waitTime)

	assert.Nil(t, res)
	assert.EqualError(t, err, expectedErr.Error())
}

func (s *StoreTestSuite) TestIsSETStatement() {
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

func (s *StoreTestSuite) TestIsUSEStatement() {
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

func (s *StoreTestSuite) TestParseSETStatement() {
	key, value, _ := parseSetStatement("SET 'key'='value'")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("SET 'key'='value';")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key'='value'    ;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key' = 'value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key'     =    'value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key    '= '		va  lue'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key' ='value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set 'key'		 ='value'    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("SET")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET 	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET 'key'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET = 'value'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET 'key'= \n'value'	")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key= \nvalue	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("set 'key'= \nvalue	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("set key= \n'value'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("set 'key= \nvalue'	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)
}

func (s *StoreTestSuite) TestParseSETStatementerror() {
	_, _, err := parseSetStatement("SET key")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: missing \"=\". Usage example: SET 'key'='value'.", err.Error())

	_, _, err = parseSetStatement("SET =")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Key and value not present. Usage example: SET 'key'='value'.", err.Error())

	_, _, err = parseSetStatement("SET key=")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Value for key not present. If you want to reset a key, use \"RESET 'key'\".", err.Error())

	_, _, err = parseSetStatement("SET =value")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Key not present. Usage example: SET 'key'='value'.", err.Error())

	_, _, err = parseSetStatement("SET ass=value=as")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: \"=\" should only appear once. Usage example: SET 'key'='value'.", err.Error())

	_, _, err = parseSetStatement("SET key=value")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Key and value must be enclosed by single quotes ''. Usage example: SET 'key'='value'.", err.Error())
}

func (s *StoreTestSuite) TestParseUSEStatement() {
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

func (s *StoreTestSuite) TestParseUSEStatementError() {
	_, _, err := parseUseStatement("USE CATALOG ;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Missing catalog name: Usage example: USE CATALOG METADATA.", err.Error())

	_, _, err = parseUseStatement("USE;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Missing database/catalog name: Usage examples: USE DB1 OR USE CATALOG METADATA.", err.Error())

	_, _, err = parseUseStatement("USE CATALOG DATABASE DB2;")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Invalid syntax for USE. Usage examples: USE CATALOG my_catalog or USE my_database", err.Error())
}

func (s *StoreTestSuite) TestParseResetStatement() {
	key, err := parseResetStatement("RESET 'key'")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("RESET 'key.key';")
	assert.Equal(s.T(), "key.key", key)
	assert.Nil(s.T(), err)

	key, err = parseResetStatement("RESET 'KEY.key';")
	assert.Equal(s.T(), "key.key", key)
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
	assert.Equal(s.T(), "key", key)
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
}

func (s *StoreTestSuite) TestParseResetStatementError() {
	key, err := parseResetStatement(" ")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Invalid syntax for RESET. Usage example: RESET 'key'.", err.Error())

	key, err = parseResetStatement("RESET key key2")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET 'key'.", err.Error())

	key, err = parseResetStatement("RESET key key2 key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET 'key'.", err.Error())

	key, err = parseResetStatement("RESET key;; key key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET 'key'.", err.Error())

	key, err = parseResetStatement("RESET key key;;; key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET 'key'.", err.Error())

	key, err = parseResetStatement("RESET key;")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Invalid syntax for RESET, key must be enclosed by single quotes ''. Usage example: RESET 'key'.", err.Error())
}

func (s *StoreTestSuite) TestProccessHttpErrors() {
	// given
	res := &http.Response{
		StatusCode: http.StatusUnauthorized,
		Body:       generateCloserFromObject(v1.NewError()),
	}

	// when
	err := processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Unauthorized. Please consider running confluent login again.", err.Error())

	// given
	title := "Invalid syntax"
	detail := "you should provide a table for select"
	statementErr := &v1.Error{Title: &title, Detail: &detail}
	res = &http.Response{
		StatusCode: http.StatusBadRequest,
		Body:       generateCloserFromObject(statementErr),
	}

	// when
	err = processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Invalid syntax: you should provide a table for select", err.Error())

	// given
	res = &http.Response{
		StatusCode: http.StatusInternalServerError,
		Body:       generateCloserFromObject(nil),
	}

	// when
	err = processHttpErrors(res, nil)

	// expect
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: received error with code \"500\" from server but could not parse it. This is not expected. Please contact support.", err.Error())

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
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement("orgId", "envId", statementName).Return(nil)

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.True(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestDeleteStatementFailsOnError() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement("orgId", "envId", statementName).Return(errors.New("test error"))

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWithCompletedStatement() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.COMPLETED,
	}
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("orgId", "envId", statement.StatementName, statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsRetryWithRunningStatement() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("orgId", "envId", statement.StatementName, statement.PageToken).Return(statementResultObj, nil).Times(5)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenPageTokenExists() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	nextPage := "https://devel.cpdev.cloud/some/results?page_token=eyJWZX"
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{Next: &nextPage},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults("orgId", "envId", statement.StatementName, statement.PageToken).Return(statementResultObj, nil)

	statementResults, err := store.FetchStatementResults(statement)
	require.NotNil(s.T(), statementResults)
	require.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWhenResultsExist() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	appOptions := types.ApplicationOptions{
		OrgResourceId: "orgId",
		EnvId:         "envId",
	}
	store := NewStore(client, mockAppController.ExitApplication, &appOptions)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{Data: &[]any{map[string]any{"op": 0}}},
	}
	client.EXPECT().GetStatementResults("orgId", "envId", statement.StatementName, statement.PageToken).Return(statementResultObj, nil)

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
		result := timeout(tc.properties)
		require.Equal(t, tc.expected, result, tc.name)
	}
}
