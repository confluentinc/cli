package lsp

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	websocket2 "github.com/sourcegraph/jsonrpc2/websocket"

	"github.com/confluentinc/cli/v4/pkg/log"
)

type WebsocketLSPClient struct {
	sync.Mutex

	baseUrl        string
	getAuthToken   func() string
	organizationId string
	environmentId  string
	handlerCh      chan *jsonrpc2.Request
	conn           *jsonrpc2.Conn
	lspClient      LspInterface
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
	close(w.handlerCh)
}

func (w *WebsocketLSPClient) client() LspInterface {
	w.Lock()
	defer w.Unlock()

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
		if lspClient, conn, err := newLSPConnection(w.baseUrl, w.getAuthToken(), w.organizationId, w.environmentId, w.handlerCh); err == nil {
			w.lspClient = lspClient
			w.conn = conn
		}
	default:
		// we need the default case here, otherwise the select/case will block until the channel has data
	}
}

func NewWebsocketClient(getAuthToken func() string, baseUrl, organizationId, environmentId string, handlerCh chan *jsonrpc2.Request) (LspInterface, error) {
	lspClient, conn, err := newLSPConnection(baseUrl, getAuthToken(), organizationId, environmentId, handlerCh)
	if err != nil {
		return nil, err
	}
	websocketClient := &WebsocketLSPClient{
		baseUrl:        baseUrl,
		getAuthToken:   getAuthToken,
		organizationId: organizationId,
		environmentId:  environmentId,
		handlerCh:      handlerCh,
		lspClient:      lspClient,
		conn:           conn,
	}
	return websocketClient, nil
}

func newLSPConnection(baseUrl, authToken, organizationId, environmentId string, handlerCh chan *jsonrpc2.Request) (LspInterface, *jsonrpc2.Conn, error) {
	stream, err := newWSObjectStream(baseUrl, authToken, organizationId, environmentId)
	if err != nil {
		log.CliLogger.Debugf("Error dialing websocket: %v\n", err)
		return nil, nil, err
	}

	lspHandler := NewLspHandler(handlerCh)

	conn := jsonrpc2.NewConn(
		context.Background(),
		stream,
		lspHandler,
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
