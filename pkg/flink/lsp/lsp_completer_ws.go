package lsp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/jsonrpc2"
	websocket2 "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/confluentinc/cli/v3/pkg/log"
)

func NewWSObjectStream(baseUrl, authToken, organizationId, environmentId string) jsonrpc2.ObjectStream {
	requestHeaders := http.Header{}
	requestHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	requestHeaders.Add("Organization-ID", organizationId)
	requestHeaders.Add("Environment-ID", environmentId)
	socketUrl := fmt.Sprintf("wss://%s/lsp", baseUrl)
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, requestHeaders)
	if err != nil {
		return nil
	}
	return websocket2.NewObjectStream(conn)
}

func NewLSPClientWS(baseUrl, authToken, organizationId, environmentId string) LSPInterface {
	stream := NewWSObjectStream(baseUrl, authToken, organizationId, environmentId)
	lspClient := &LSPClient{
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
