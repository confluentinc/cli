package lsp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	websocket2 "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
)

func NewWSObjectStream(store types.StoreInterface) jsonrpc2.ObjectStream {
	requestHeaders := http.Header{}
	requestHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", store.GetAuthToken()))
	requestHeaders.Add("Organization-ID", store.GetOrganizationId())
	requestHeaders.Add("Environment-ID", store.GetEnvironmentId())
	socketUrl := "ws://localhost:8000/lsp"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, requestHeaders)
	if err != nil {
		return nil
	}
	return websocket2.NewObjectStream(conn)
}

func NewLSPClientWS(store types.StoreInterface) LSPInterface {
	stream := NewWSObjectStream(store)
	lspClient := &LSPClient{
		store: store,
		conn: jsonrpc2.NewConn(
			context.Background(),
			stream,
			noopHandler{},
			nil,
		),
	}

	lspInitParams, err := lspClient.Initialize()
	if err != nil {
		log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
		return nil
	}

	log.CliLogger.Trace("LSP init params: ", lspInitParams)
	err = lspClient.DidOpen()
	if err != nil {
		log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
		return nil
	}

	return lspClient
}
