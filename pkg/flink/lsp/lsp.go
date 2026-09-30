package lsp

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/sourcegraph/go-lsp"

	"github.com/confluentinc/cli/v4/pkg/flink/types"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type LspInterface interface {
	Initialize() (*lsp.InitializeResult, error)
	DidOpen() (lsp.DocumentURI, error)
	DidChange(newText string) error
	DidChangeConfiguration(settings any) error
	Completion(position lsp.Position) (lsp.CompletionList, error)
	ShutdownAndExit()
	CurrentDocumentUri() string
}

type LSPClient struct {
	conn        types.JSONRpcConn
	documentURI *lsp.DocumentURI
}

type CliContext struct {
	AuthToken           string
	Catalog             string
	Database            string
	ComputePoolId       string
	LspDocumentUri      string
	StatementProperties map[string]string
}

func NewLSPClientImpl(conn types.JSONRpcConn) *LSPClient {
	return &LSPClient{
		conn: conn,
	}
}

func (c *LSPClient) Initialize() (*lsp.InitializeResult, error) {
	log.CliLogger.Debugf("Initialize called")
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

func (c *LSPClient) DidOpen() (lsp.DocumentURI, error) {
	log.CliLogger.Debugf("DidOpen called")

	var resp interface{}

	documentURI := lsp.DocumentURI("temp_session_" + uuid.New().String() + ".sql")

	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:  documentURI,
			Text: "",
		},
	}

	if c.conn == nil {
		return documentURI, errors.New("connection to LSP server not established/nil")
	}

	err := c.conn.Call(context.Background(), "textDocument/didOpen", didOpenParams, &resp)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return documentURI, err
	}
	c.documentURI = &documentURI
	return documentURI, nil
}

func (c *LSPClient) DidChange(newText string) error {
	log.CliLogger.Debugf("DidChange called")
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

	err := c.conn.Notify(context.Background(), "textDocument/didChange", didchangeParams)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	return nil
}

func (c *LSPClient) DidChangeConfiguration(settings any) error {
	log.CliLogger.Debugf("DidChangeConfiguration called")
	if c.conn == nil {
		return errors.New("connection to LSP server not established/nil")
	}

	didChangeConfigParams := lsp.DidChangeConfigurationParams{
		Settings: settings,
	}

	err := c.conn.Notify(context.Background(), "workspace/didChangeConfiguration", didChangeConfigParams)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return err
	}
	return nil
}

func (c *LSPClient) Completion(position lsp.Position) (lsp.CompletionList, error) {
	log.CliLogger.Debugf("Completion called")
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

	err := c.conn.Call(context.Background(), "textDocument/completion", completionParams, &resp)

	if err != nil {
		log.CliLogger.Debugf("Error sending request: %v\n", err)
		return lsp.CompletionList{}, err
	}

	return resp, nil
}

func (c *LSPClient) ShutdownAndExit() {
	log.CliLogger.Debugf("ShutdownAndExit called")
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
		log.CliLogger.Debugf("Error exiting lsp server: %v\n", err)
	}
}

// CurrentDocumentUri returns the currently opened document. Obs.: This CLI LSP interface currently only supports one document at a time.
func (c *LSPClient) CurrentDocumentUri() string {
	if c.documentURI == nil {
		return ""
	}

	return string(*c.documentURI)
}
