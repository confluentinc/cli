package autocomplete

import (
	"context"
	"errors"
	"fmt"
	"github.com/confluentinc/cli/v3/pkg/flink/internal/utils"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
	websocket2 "github.com/sourcegraph/jsonrpc2/websocket"
	"net/http"
	"time"

	"github.com/confluentinc/flink-sql-language-service/pkg/api"
	prompt "github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
)

const debounceTimeout = 200 * time.Millisecond

type LSPClientWS struct {
	conn        types.JSONRpcConn
	documentURI *lsp.DocumentURI
	store       types.StoreInterface
	debouncer   utils.Debouncer[[]prompt.Suggest]
}

func (c *LSPClientWS) LSPCompleter(in prompt.Document) []prompt.Suggest {
	textBeforeCursor := in.TextBeforeCursor()
	startOfPreviousWord := in.FindStartOfPreviousWord()

	if textBeforeCursor == "" {
		return nil
	}

	return c.debouncer.Debounce(func() []prompt.Suggest {
		err := c.didChange(in.Text)
		if err != nil {
			log.CliLogger.Debugf("Error sending didChange lsp request: %v\n", err)
			return nil
		}

		position := lsp.Position{
			Line:      0,
			Character: len(textBeforeCursor),
		}

		completionList, err := c.completion(position)
		if err != nil {
			log.CliLogger.Debugf("Error sending completion lsp request: %v\n", err)
			return nil
		}

	return lspCompletionsToSuggests(completionList.Items, in.GetWordBeforeCursor(), startOfPreviousWord)
}

func (c *LSPClientWS) initialize() (*lsp.InitializeResult, error) {
	var resp lsp.InitializeResult

	initializeParams := lsp.InitializeParams{}

	if c.conn == nil {
		return nil, errors.New("connection to LSP server not established/nil")
	}

	err := c.conn.Call(context.Background(), "initialize", initializeParams, &resp)

	if err != nil {
		log.CliLogger.Debugf("Error initializing LSP: %v\n", err)
	}

	return &resp, err
}

func (c *LSPClientWS) didOpen() error {
	var resp interface{}

	documentURI := lsp.DocumentURI("temp_session_" + uuid.New().String() + ".sql")

	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:  documentURI,
			Text: "",
		},
	}

	if c.conn == nil {
		return errors.New("connection to LSP server not established/nil")
	}

	err := c.conn.Call(context.Background(), "textDocument/didOpen", didOpenParams, &resp)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	c.documentURI = &documentURI
	return nil
}

func (c *LSPClientWS) didChange(newText string) error {
	var resp interface{}

	if c.conn == nil || c.documentURI == nil {
		return errors.New("connection to LSP server not established/nil")
	}

	didchangeParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: *c.documentURI,
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{Text: newText},
		},
	}

	err := c.conn.Call(context.Background(), "textDocument/didChange", didchangeParams, &resp)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	return nil
}

func (c *LSPClientWS) completion(position lsp.Position) (lsp.CompletionList, error) {
	var resp lsp.CompletionList

	if c.conn == nil || c.documentURI == nil {
		return lsp.CompletionList{}, errors.New("connection to LSP server not established/nil")
	}

	completionParams := lsp.CompletionParams{TextDocumentPositionParams: lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: *c.documentURI,
		},
		Position: position,
	}}
	cliCtx := api.CliContext{
		AuthToken:     c.store.GetAuthToken(),
		Catalog:       c.store.GetCurrentCatalog(),
		Database:      c.store.GetCurrentDatabase(),
		ComputePoolId: c.store.GetComputePool(),
	}

	err := c.conn.Call(context.Background(), "textDocument/completion", completionParams, &resp, jsonrpc2.Meta(cliCtx))

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return lsp.CompletionList{}, err
	}

	return resp, nil
}

func (c *LSPClientWS) ShutdownAndExit() {
	if c.conn == nil {
		return
	}

	err := c.conn.Call(context.Background(), "shutdown", nil, nil)

	if err != nil {
		log.CliLogger.Debugf("Error shutting down lsp server: %v\n", err)
		return
	}

	err = c.conn.Call(context.Background(), "exit", nil, nil)

	if err != nil {
		log.CliLogger.Debugf("Error existing lsp server: %v\n", err)
	}
}

func NewLSPClientWS(store types.StoreInterface) LSPClientInterface {
	lspClient := &LSPClientWS{
		store:     store,
		debouncer: utils.NewDebouncer[[]prompt.Suggest](nil, debounceTimeout),
	}

	requestHeaders := http.Header{}
	requestHeaders.Add("Authorization", fmt.Sprintf("Bearer %s", store.GetAuthToken()))
	requestHeaders.Add("Organization-ID", store.GetOrganizationId())
	requestHeaders.Add("Environment-ID", store.GetEnvironmentId())
	socketUrl := "ws://localhost:8000/lsp"
	conn, _, err := websocket.DefaultDialer.Dial(socketUrl, requestHeaders)

	if err == nil {
		stream := websocket2.NewObjectStream(conn)
		jsonRpcConn := jsonrpc2.NewConn(
			context.Background(),
			stream,
			noopHandler{},
			nil,
		)
		lspClient.conn = jsonRpcConn

		lspInitParams, err := lspClient.initialize()
		if err != nil {
			log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
			return nil
		}

		log.CliLogger.Trace("LSP init params: ", lspInitParams)
		err = lspClient.didOpen()
		if err != nil {
			log.CliLogger.Debugf("Error opening lsp connection: %v\n", err)
			return nil
		}
	}

	return lspClient
}
