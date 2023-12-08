package lsp

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/sourcegraph/go-lsp"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/confluentinc/cli/v3/pkg/flink/test/mock"
)

func TestLSPIntialize(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	_, err := lspClient.Initialize()
	require.NoError(t, err)
}

func TestLSPIntializeCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	_, err := lspClient.Initialize()
	require.Error(t, err)
}

func TestLSPIntializeNoConnErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: nil}
	_, err := lspClient.Initialize()
	require.Error(t, err)
}

func TestLSPdidOpen(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	lspClient := &LSPClient{conn: conn}
	require.Nil(t, lspClient.documentURI)
	err := lspClient.DidOpen()
	require.NotNil(t, lspClient.documentURI)
	require.NoError(t, err)
}

func TestLSPdidOpenCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)

	lspClient := &LSPClient{conn: conn}
	require.Nil(t, lspClient.documentURI)
	err := lspClient.DidOpen()
	require.Nil(t, lspClient.documentURI)
	require.Error(t, err)
}

func TestLSPdidOpenNoConnErr(t *testing.T) {
	lspClient := &LSPClient{conn: nil}
	err := lspClient.DidOpen()
	require.Error(t, err)
}

func TestLSPdidChange(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Notify(gomock.Any(), "textDocument/didChange", gomock.Any()).Return(nil).Times(1)

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	err := lspClient.DidChange("some text")
	require.NoError(t, err)
}

func TestLSPdidChangeCallErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Notify(gomock.Any(), "textDocument/didChange", gomock.Any()).Return(errors.New("some error")).Times(1)

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	err := lspClient.DidChange("some text")
	require.Error(t, err)
}

func TestLSPdidChangeNoConnErr(t *testing.T) {
	lspClient := &LSPClient{conn: nil}
	err := lspClient.DidChange("some text")
	require.Error(t, err)
}

func TestLSPCompletion(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	Completion, err := lspClient.Completion(lsp.Position{})

	require.NoError(t, err)
	require.NotNil(t, Completion)
}

func TestLSPCompletionCallErr(t *testing.T) {
	conn := mock.NewMockJSONRpcConn(gomock.NewController(t))
	conn.EXPECT().Call(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("some error")).Times(1)
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: conn}
	Completion, err := lspClient.Completion(lsp.Position{})

	require.Error(t, err)
	require.Equal(t, lsp.CompletionList{}, Completion)
}

func TestLSPCompletionNoConnErr(t *testing.T) {
	uri := lsp.DocumentURI("file:///test.sql")

	lspClient := &LSPClient{documentURI: &uri, conn: nil}
	_, err := lspClient.Completion(lsp.Position{})
	require.Error(t, err)
}
