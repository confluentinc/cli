package autocomplete

import (
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/sourcegraph/go-lsp"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/flink/internal/store"
	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
)

func TestLSPIntialize(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	_, err := lspClient.initialize()
	require.NoError(t, err)
}

func TestLSPIntializeCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	_, err := lspClient.initialize()
	require.Error(t, err)
}

func TestLSPIntializeNoConnErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: nil}
	_, err := lspClient.initialize()
	require.Error(t, err)
}

func TestLSPdidOpen(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	lspClient := &LSPClient{conn: conn}
	require.Nil(t, lspClient.documentURI)
	err := lspClient.didOpen()
	require.NotNil(t, lspClient.documentURI)
	require.NoError(t, err)
}

func TestLSPdidOpenCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)

	lspClient := &LSPClient{conn: conn}
	require.Nil(t, lspClient.documentURI)
	err := lspClient.didOpen()
	require.Nil(t, lspClient.documentURI)
	require.Error(t, err)
}

func TestLSPdidOpenNoConnErr(t *testing.T) {
	lspClient := &LSPClient{conn: nil}
	err := lspClient.didOpen()
	require.Error(t, err)
}

func TestLSPdidChange(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	err := lspClient.didChange("some text")
	require.NoError(t, err)
}

func TestLSPdidChangeCallErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	err := lspClient.didChange("some text")
	require.Error(t, err)
}

func TestLSPdidChangeNoConnErr(t *testing.T) {
	lspClient := &LSPClient{conn: nil}
	err := lspClient.didChange("some text")
	require.Error(t, err)
}

func TestLSPCompletion(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")
	s := store.NewStore(mock.NewFakeFlinkGatewayClient(), func() {}, &types.ApplicationOptions{}, func() error { return nil })

	lspClient := &LSPClient{documentURI: &uri, conn: conn, store: s}
	completion, err := lspClient.completion(lsp.Position{})

	require.NoError(t, err)
	require.NotNil(t, completion)
}

func TestLSPCompletionCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")
	s := store.NewStore(mock.NewFakeFlinkGatewayClient(), func() {}, &types.ApplicationOptions{}, func() error { return nil })

	lspClient := &LSPClient{documentURI: &uri, conn: conn, store: s}
	completion, err := lspClient.completion(lsp.Position{})

	require.Error(t, err)
	require.Equal(t, lsp.CompletionList{}, completion)
}

func TestLSPCompletionNoConnErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: nil}
	_, err := lspClient.completion(lsp.Position{})
	require.Error(t, err)
}

func TestNewLSPClient(t *testing.T) {
	lspClient := NewLSPClient(nil).(*LSPClient)
	require.NotNil(t, lspClient)

	isRunning := waitForConditionWithTimeout(func() bool { return lspClient.conn != nil }, 3*time.Second)
	require.True(t, isRunning)
}
