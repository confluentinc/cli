package lsp

import (
	"context"
	"crypto/tls"
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

	baseUrl         string
	getAuthToken    func() string
	organizationId  string
	environmentId   string
	handlerCh       chan *jsonrpc2.Request
	conn            *jsonrpc2.Conn
	lspClient       LspInterface
	tlsClientConfig *tls.Config
}

func (w *WebsocketLSPClient) Initialize() (*lsp.InitializeResult, error) {
	return w.client().Initialize()
}

func (w *WebsocketLSPClient) DidOpen() (lsp.DocumentURI, error) {
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

func (w *WebsocketLSPClient) CurrentDocumentUri() string {
	return w.client().CurrentDocumentUri()
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
		log.CliLogger.Debugf("Language service websocket disconnected, reconnecting")
		// this shouldn't happen, but if we are connected do nothing
		if isConnected {
			break
		}

		// we only update client and conn if there was no error, otherwise we leave them as is
		if lspClient, conn, err := NewLSPConnection(w.baseUrl, w.getAuthToken(), w.organizationId, w.environmentId, w.tlsClientConfig, w.handlerCh); err == nil {
			w.lspClient = lspClient
			w.conn = conn
			_, err := InitLspClient(w.lspClient)
			if err != nil {
				log.CliLogger.Debugf("Error initializing lsp connection: %v\n", err)
			}
		}
	default:
		// we need the default case here, otherwise the select/case will block until the channel has data
	}
}

func NewInitializedLspClient(getAuthToken func() string, baseUrl, organizationId, environmentId string, tlsClientConfig *tls.Config, handlerCh chan *jsonrpc2.Request) (LspInterface, string, error) {
	client, _, err := NewLSPClient(getAuthToken, baseUrl, organizationId, environmentId, tlsClientConfig, handlerCh)
	if err != nil {
		return nil, "", err
	}

	docUri, err := InitLspClient(client)

	if err != nil {
		return nil, docUri, err
	}

	return client, docUri, nil
}

func InitLspClient(client LspInterface) (string, error) {
	_, err := client.Initialize()
	if err != nil {
		return "", err
	}
	log.CliLogger.Debugf("Language service intialized")

	docUri, err := client.DidOpen()
	log.CliLogger.Debugf("Language service opened document: %s", docUri)
	if err != nil {
		return "", err
	}
	return string(docUri), nil
}

func NewLSPClient(getAuthToken func() string, baseUrl, organizationId, environmentId string, tlsClientConfig *tls.Config, handlerCh chan *jsonrpc2.Request) (*WebsocketLSPClient, *jsonrpc2.Conn, error) {
	lspClient, conn, err := NewLSPConnection(baseUrl, getAuthToken(), organizationId, environmentId, tlsClientConfig, handlerCh)
	if err != nil {
		return nil, conn, err
	}

	websocketClient := &WebsocketLSPClient{
		baseUrl:         baseUrl,
		getAuthToken:    getAuthToken,
		organizationId:  organizationId,
		environmentId:   environmentId,
		handlerCh:       handlerCh,
		lspClient:       lspClient,
		conn:            conn,
		tlsClientConfig: tlsClientConfig,
	}

	return websocketClient, websocketClient.conn, nil
}

func NewLSPConnection(baseUrl, authToken, organizationId, environmentId string, tlsClientConfig *tls.Config, handlerCh chan *jsonrpc2.Request) (*LSPClient, *jsonrpc2.Conn, error) {
	stream, err := NewWSObjectStream(baseUrl, authToken, organizationId, environmentId, tlsClientConfig)
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
	lspClient := NewLSPClientImpl(conn)

	return lspClient, conn, nil
}

func NewWSObjectStream(socketUrl, authToken, organizationId, environmentId string, tlsClientConfig *tls.Config) (jsonrpc2.ObjectStream, error) {
	requestHeaders := http.Header{}
	requestHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", authToken))
	requestHeaders.Add("Organization-ID", organizationId)
	requestHeaders.Add("Environment-ID", environmentId)
	dialer := &websocket.Dialer{
		TLSClientConfig: tlsClientConfig,
	}
	conn, _, err := dialer.Dial(socketUrl, requestHeaders)
	if err != nil {
		return nil, err
	}
	return websocket2.NewObjectStream(conn), nil
}
