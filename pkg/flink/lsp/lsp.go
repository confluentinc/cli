package lsp

import (
	"context"
	"errors"
	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/google/uuid"
	"github.com/sourcegraph/go-lsp"
)

type LSPInterface interface {
	Initialize() (*lsp.InitializeResult, error)
	DidOpen() error
	DidChange(newText string) error
	DidChangeConfiguration(settings any) error
	Completion(position lsp.Position) (lsp.CompletionList, error)
	ShutdownAndExit()
}

type LSPClient struct {
	conn        types.JSONRpcConn
	documentURI *lsp.DocumentURI
}

func (c *LSPClient) Initialize() (*lsp.InitializeResult, error) {
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

func (c *LSPClient) DidOpen() error {
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

func (c *LSPClient) DidChange(newText string) error {
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
