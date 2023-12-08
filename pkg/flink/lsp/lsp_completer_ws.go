package lsp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	websocket2 "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/confluentinc/cli/v3/pkg/log"
)

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

type WebsocketLSPClient struct {
	baseUrl        string
	getAuthToken   func() string
	organizationId string
	environmentId  string

	conn      *jsonrpc2.Conn
	lspClient LSPInterface
}

func (w *WebsocketLSPClient) Initialize() (*lsp.InitializeResult, error) {
	return w.client().Initialize()
}

func (w *WebsocketLSPClient) DidOpen() error {
	return w.client().DidOpen()
}

func (w *WebsocketLSPClient) DidChange(newText string) error {
	return w.client().DidChange(newText)
}

func (w *WebsocketLSPClient) DidChangeConfiguration(settings any) error {
	return w.client().DidChangeConfiguration(settings)
}

func (w *WebsocketLSPClient) Completion(position lsp.Position) (lsp.CompletionList, error) {
	return w.client().Completion(position)
}

func (w *WebsocketLSPClient) ShutdownAndExit() {
	w.client().ShutdownAndExit()
}

func (w *WebsocketLSPClient) client() LSPInterface {
	w.refreshWebsocketConnection()
	return w.lspClient
}

func (w *WebsocketLSPClient) refreshWebsocketConnection() {
	select {
	case _, isConnected := <-w.conn.DisconnectNotify():
		// this shouldn't happen, but if we are connected do nothing
		if isConnected {
			break
		}

		// we only update client and conn if there was no error, otherwise we leave them as is
		if lspClient, conn, err := newLSPConnection(w.baseUrl, w.getAuthToken(), w.organizationId, w.environmentId); err == nil {
			w.lspClient = lspClient
			w.conn = conn
		}
	default:
		// we need the default case here, otherwise the select/case will block until the channel has data
	}
}

func NewLSPClientWS(getAuthToken func() string, baseUrl, organizationId, environmentId string) LSPInterface {
	lspClient, conn, err := newLSPConnection(baseUrl, getAuthToken(), organizationId, environmentId)
	if err != nil {
		return nil
	}
	websocketClient := &WebsocketLSPClient{
		baseUrl:        baseUrl,
		getAuthToken:   getAuthToken,
		organizationId: organizationId,
		environmentId:  environmentId,
		lspClient:      lspClient,
		conn:           conn,
	}
	return websocketClient
}

func newLSPConnection(baseUrl, authToken, organizationId, environmentId string) (LSPInterface, *jsonrpc2.Conn, error) {
	stream, err := newWSObjectStream(baseUrl, authToken, organizationId, environmentId)
	if err != nil {
		log.CliLogger.Debugf("Error dialing websocket: %v\n", err)
		return nil, nil, err
	}

	conn := jsonrpc2.NewConn(
		context.Background(),
		stream,
		noopHandler{},
		nil,
	)
	lspClient := NewLSPClient(conn)

	lspInitParams, err := lspClient.Initialize()
	if err != nil {
		log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
		return nil, nil, err
	}

	log.CliLogger.Trace("LSP init params: ", lspInitParams)
	err = lspClient.DidOpen()
	if err != nil {
		log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
		return nil, nil, err
	}

	return lspClient, conn, nil
}

func newWSObjectStream(socketUrl, authToken, organizationId, environmentId string) (jsonrpc2.ObjectStream, error) {
	requestHeaders := http.Header{}
	requestHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	requestHeaders.Add("Organization-ID", organizationId)
	requestHeaders.Add("Environment-ID", environmentId)
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, requestHeaders)
	if err != nil {
		return nil, err
	}
	return websocket2.NewObjectStream(conn), nil
}
