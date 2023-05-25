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

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/flink-gateway/v1alpha1"

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
	client := NewGatewayClient("envId", "orgResourceId", "kafkaClusterId", "computePoolId", func() string { return "authToken" }, nil)
	mockAppController := mock.NewMockApplicationControllerInterface(gomock.NewController(t))
	s := NewStore(client, mockAppController.ExitApplication, nil).(*Store)

	result, err := s.ProcessLocalStatement("SET foo=bar;")
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
	s := &Store{
		client: client,
	}

	// Test case 1: Statement is not pending
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "COMPLETED",
		},
	}
	httpRes := http.Response{StatusCode: http.StatusOK}
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObj, &httpRes, nil).Times(1)

	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, time.Duration(10))
	assert.Nil(t, err)
	assert.NotNil(t, processedStatement)
	assert.Equal(t, types.NewProcessedStatement(statementObj), processedStatement)
}

func TestWaitForPendingTimesout(t *testing.T) {
	statementName := "statementName"
	timeout := time.Duration(10) * time.Millisecond

	httpRes := http.Response{StatusCode: http.StatusOK}
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	s := &Store{
		client: client,
	}

	// Test case 2: Statement is pending
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObj, &httpRes, nil).AnyTimes()
	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, timeout)

	assert.EqualError(t, err, fmt.Sprintf("Error: Statement is still pending after %f seconds. \n\nIf you want to increase the timeout for the client, you can run \"SET table.results-timeout=1200;\" to adjust the maximum timeout in seconds.", timeout.Seconds()))
	assert.Nil(t, processedStatement)
}

func TestWaitForPendingEventuallyCompletes(t *testing.T) {
	statementName := "statementName"

	httpRes := http.Response{StatusCode: http.StatusOK}
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	s := &Store{
		client: client,
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
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObj, &httpRes, nil).Times(3)
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObjCompleted, &httpRes, nil).Times(1)

	processedStatement, err := s.waitForPendingStatement(context.Background(), statementName, time.Duration(10)*time.Second)
	assert.Nil(t, err)
	assert.NotNil(t, processedStatement)
	assert.Equal(t, types.NewProcessedStatement(statementObjCompleted), processedStatement)
}

func TestWaitForPendingStatementErrors(t *testing.T) {
	statementName := "statementName"
	waitTime := time.Millisecond * 1
	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	s := &Store{
		client: client,
	}
	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "COMPLETED",
		},
	}

	expectedErr := errors.New("couldn't get statement!")
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObj, nil, expectedErr).Times(1)
	_, err := s.waitForPendingStatement(context.Background(), statementName, waitTime)
	assert.EqualError(t, err, "Error: "+expectedErr.Error())
}

func TestCancelPendingStatement(t *testing.T) {
	statementName := "statementName"
	waitTime := time.Second * 1
	ctx, cancelFunc := context.WithCancel(context.Background())

	client := mock.NewMockGatewayClientInterface(gomock.NewController(t))
	s := &Store{
		client: client,
	}

	statementObj := v1.SqlV1alpha1Statement{
		Status: &v1.SqlV1alpha1StatementStatus{
			Phase: "PENDING",
		},
	}
	httpRes := http.Response{StatusCode: http.StatusOK}

	expectedErr := &types.StatementError{Msg: "Result retrieval aborted. Statement will be deleted."}
	client.EXPECT().GetStatement(gomock.Any(), statementName).Return(statementObj, &httpRes, nil).AnyTimes()

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
	assert.True(s.T(), true, statementStartsWithOp("SET", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("SET key=value", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET   ", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    set   ", configOpSet))
	assert.True(s.T(), true, statementStartsWithOp("    SET key=value", configOpSet))

	assert.False(s.T(), false, statementStartsWithOp("SETting", configOpSet))
	assert.False(s.T(), false, statementStartsWithOp("", configOpSet))
	assert.False(s.T(), false, statementStartsWithOp("should be false", configOpSet))
	assert.False(s.T(), false, statementStartsWithOp("USE", configOpSet))
	assert.False(s.T(), false, statementStartsWithOp("SETTING", configOpSet))
}

func (s *StoreTestSuite) TestIsUSEStatement() {
	assert.True(s.T(), statementStartsWithOp("USE", configOpUse))
	assert.True(s.T(), statementStartsWithOp("USE catalog", configOpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", configOpUse))
	assert.True(s.T(), statementStartsWithOp("use CATALOG cat", configOpUse))
	assert.True(s.T(), statementStartsWithOp("USE   ", configOpUse))
	assert.True(s.T(), statementStartsWithOp("use   ", configOpUse))
	assert.True(s.T(), statementStartsWithOp("USE CATALOG cat", configOpUse))

	assert.False(s.T(), statementStartsWithOp("SET", configOpUse))
	assert.False(s.T(), statementStartsWithOp("USES", configOpUse))
	assert.False(s.T(), statementStartsWithOp("", configOpUse))
	assert.False(s.T(), statementStartsWithOp("should be false", configOpUse))
}

func (s *StoreTestSuite) TestIsResetStatement() {
	assert.True(s.T(), true, statementStartsWithOp("RESET", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET   ", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("reset   ", configOpReset))
	assert.True(s.T(), true, statementStartsWithOp("RESET key=value", configOpReset))

	assert.False(s.T(), false, statementStartsWithOp("RESETting", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("USE", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("RESETTING", configOpReset))
}

func (s *StoreTestSuite) TestIsExitStatement() {
	assert.True(s.T(), true, statementStartsWithOp("EXIT", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("EXIT ;", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit   ;", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exiT   ", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("Exit   ", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("eXit   ", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit", configOpExit))
	assert.True(s.T(), true, statementStartsWithOp("exit ", configOpExit))

	assert.False(s.T(), false, statementStartsWithOp("exits", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("should be false", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("exitt;", configOpReset))
	assert.False(s.T(), false, statementStartsWithOp("exi", configOpReset))
}

func (s *StoreTestSuite) TestParseSETStatement() {
	key, value, _ := parseSetStatement("SET key=value")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("SET key=value;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key=value    ;")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key = value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key     =    value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key= value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key =value    ")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)

	key, value, _ = parseSetStatement("set key		 =value    ")
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

	key, value, _ = parseSetStatement("sET key	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET = value	")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseSetStatement("sET key= \nvalue	")
	assert.Equal(s.T(), "key", key)
	assert.Equal(s.T(), "value", value)
}

func (s *StoreTestSuite) TestParseSETStatementerror() {
	_, _, err := parseSetStatement("SET key")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: missing \"=\". Usage example: SET key=value.", err.Error())

	_, _, err = parseSetStatement("SET =")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Key and value not present. Usage example: SET key=value.", err.Error())

	_, _, err = parseSetStatement("SET key=")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Value for key not present. If you want to reset a key, use \"RESET key\".", err.Error())

	_, _, err = parseSetStatement("SET =value")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Key not present. Usage example: SET key=value.", err.Error())

	_, _, err = parseSetStatement("SET ass=value=as")
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: \"=\" should only appear once. Usage example: SET key=value.", err.Error())
}

func (s *StoreTestSuite) TestParseUSEStatement() {
	key, value, _ := parseUseStatement("USE CATALOG c;")
	assert.Equal(s.T(), configKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value, _ = parseUseStatement("use   catalog   \nc   ")
	assert.Equal(s.T(), configKeyCatalog, key)
	assert.Equal(s.T(), "c", value)

	key, value, _ = parseUseStatement("use   catalog     ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("catalog   c")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("use     db   ")
	assert.Equal(s.T(), configKeyDatabase, key)
	assert.Equal(s.T(), "db", value)

	key, value, _ = parseUseStatement("dAtaBaSe  db   ")
	assert.Equal(s.T(), "", key)
	assert.Equal(s.T(), "", value)

	key, value, _ = parseUseStatement("use     \ndatabase_name   ")
	assert.Equal(s.T(), configKeyDatabase, key)
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
	key, err := parseResetStatement("RESET key")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("RESET key.key;")
	assert.Equal(s.T(), "key.key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("RESET KEY.key;")
	assert.Equal(s.T(), "key.key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reset key    ;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reset key   ")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reset key;;;;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reset")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("RESET")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reSET 	")
	assert.Equal(s.T(), "", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("reSET key	")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("resET KEY ")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)

	key, _ = parseResetStatement("resET key;;;")
	assert.Equal(s.T(), "key", key)
	assert.Nil(s.T(), err)
}

func (s *StoreTestSuite) TestParseResetStatementError() {
	key, err := parseResetStatement(" ")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: Invalid syntax for RESET. Usage example: RESET key.", err.Error())

	key, err = parseResetStatement("RESET key key2")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET key.", err.Error())

	key, err = parseResetStatement("RESET key key2 key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET key.", err.Error())

	key, err = parseResetStatement("RESET key;; key key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET key.", err.Error())

	key, err = parseResetStatement("RESET key key;;; key3")
	assert.Equal(s.T(), "", key)
	assert.NotNil(s.T(), err)
	assert.Equal(s.T(), "Error: too many keys for RESET provided. Usage example: RESET key.", err.Error())
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
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement(gomock.Any(), statementName).Return(nil, nil)

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.True(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestDeleteStatementFailsOnError() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	store := NewStore(client, mockAppController.ExitApplication, nil).(*Store)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement(gomock.Any(), statementName).Return(nil, errors.New("test error"))

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestDeleteStatementFailsOn404() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statementName := "TEST_STATEMENT"
	client.EXPECT().DeleteStatement(gomock.Any(), statementName).Return(&http.Response{StatusCode: http.StatusNotFound}, nil)

	wasStatementDeleted := store.DeleteStatement(statementName)
	require.False(s.T(), wasStatementDeleted)
}

func (s *StoreTestSuite) TestFetchResultsNoRetryWithCompletedStatement() {
	ctrl := gomock.NewController(s.T())
	defer ctrl.Finish()

	// create objects
	client := mock.NewMockGatewayClientInterface(ctrl)
	mockAppController := mock.NewMockApplicationControllerInterface(ctrl)
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.COMPLETED,
	}
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults(gomock.Any(), statement.StatementName, statement.PageToken).Return(statementResultObj, nil, nil)

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
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults(gomock.Any(), statement.StatementName, statement.PageToken).Return(statementResultObj, nil, nil).Times(5)

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
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	nextPage := "https://devel.cpdev.cloud/some/results?page_token=eyJWZX"
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{Next: &nextPage},
		Results:  &v1.SqlV1alpha1StatementResultResults{},
	}
	client.EXPECT().GetStatementResults(gomock.Any(), statement.StatementName, statement.PageToken).Return(statementResultObj, nil, nil)

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
	store := NewStore(client, mockAppController.ExitApplication, nil)

	statement := types.ProcessedStatement{
		StatementName: "TEST_STATEMENT",
		Status:        types.RUNNING,
	}
	op := int32(0)
	statementResultObj := v1.SqlV1alpha1StatementResult{
		Metadata: v1.ResultListMeta{},
		Results:  &v1.SqlV1alpha1StatementResultResults{Data: &[]v1.SqlV1alpha1ResultItem{{Op: &op}}},
	}
	client.EXPECT().GetStatementResults(gomock.Any(), statement.StatementName, statement.PageToken).Return(statementResultObj, nil, nil)

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
				configKeyResultsTimeout: "10", // timeout in seconds
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
				configKeyResultsTimeout: "abc", // invalid duration
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
