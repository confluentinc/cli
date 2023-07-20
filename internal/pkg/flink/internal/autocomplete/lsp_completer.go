package autocomplete

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/confluentinc/flink-sql-language-service/pkg/server/tcp"
	prompt "github.com/confluentinc/go-prompt"
	"github.com/google/uuid"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

type LSPClientInterface interface {
	LSPCompleter(in prompt.Document) []prompt.Suggest
	ShutdownAndExit()
}

type LSPClient struct {
	conn        *jsonrpc2.Conn
	documentURI lsp.DocumentURI
}

func (c *LSPClient) LSPCompleter(in prompt.Document) []prompt.Suggest {
	c.didChange(in.Text)
	textBeforeCursor := in.TextBeforeCursor()

	position := lsp.Position{
		Line:      0,
		Character: len(textBeforeCursor),
	}

	completions := []lsp.CompletionItem{}
	if textBeforeCursor != "" {
		completions = c.completion(position)
	}

	return lspCompletionsToSuggests(completions)
}

func (c *LSPClient) didChange(newText string) {
	var resp interface{}

	didchangeParams := lsp.DidChangeTextDocumentParams{
		TextDocument: lsp.VersionedTextDocumentIdentifier{
			TextDocumentIdentifier: lsp.TextDocumentIdentifier{
				URI: c.documentURI,
			},
		},
		ContentChanges: []lsp.TextDocumentContentChangeEvent{
			{Text: newText},
		},
	}

	if c.conn == nil {
		return
	}

	err := c.conn.Call(context.Background(), "textDocument/didChange", didchangeParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}
}

func (c *LSPClient) completion(position lsp.Position) []lsp.CompletionItem {
	var resp []lsp.CompletionItem

	completionParams := lsp.CompletionParams{TextDocumentPositionParams: lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{
			URI: c.documentURI,
		},
		Position: position,
	}}

	if c.conn == nil {
		return []lsp.CompletionItem{}
	}

	err := c.conn.Call(context.Background(), "textDocument/completion", completionParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}

	return resp
}

func (c *LSPClient) initialize() (*lsp.InitializeResult, error) {
	var resp lsp.InitializeResult

	initializeParams := lsp.InitializeParams{}

	if c.conn == nil {
		return nil, errors.New("connection to LSP server not established/nil")
	}

	err := c.conn.Call(context.Background(), "textDocument/initialize", initializeParams, &resp)

	if err != nil {
		fmt.Printf("Error initializing LSP: %v\n", err)
	}

	return &resp, err
}

func (c *LSPClient) didOpen() {
	var resp interface{}

	c.documentURI = lsp.DocumentURI("temp_session_" + uuid.New().String() + ".sql")
	didOpenParams := lsp.DidOpenTextDocumentParams{
		TextDocument: lsp.TextDocumentItem{
			URI:  c.documentURI,
			Text: "",
		},
	}

	if c.conn == nil {
		return
	}

	err := c.conn.Call(context.Background(), "textDocument/initialize", didOpenParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}
}

func (c *LSPClient) ShutdownAndExit() {
	if c.conn == nil {
		return
	}

	err := c.conn.Call(context.Background(), "textDocument/shutdown", nil, nil)

	if err != nil {
		fmt.Printf("Error shutting down lsp server: %v\n", err)
		return
	}

	err = c.conn.Call(context.Background(), "textDocument/exit", nil, nil)

	if err != nil {
		fmt.Printf("Error existing lsp server: %v\n", err)
	}
}

func NewLSPClient() LSPClientInterface {
	lspClient := LSPClient{}

	go func() {
		for port := 49152; port <= 65535; port++ {
			address := fmt.Sprintf("localhost:%d", port)
			lspServer := tcp.NewServer(address)

			go lspServer.Serve()

			waitForConditionWithTimeout(lspServer.IsRunning, 1*time.Second)
			conn, err := net.Dial("tcp", address)

			if err == nil {
				stream := jsonrpc2.NewBufferedStream(conn, jsonrpc2.VSCodeObjectCodec{})
				jsonRpcConn := jsonrpc2.NewConn(
					context.Background(),
					stream,
					noopHandler{},
					nil,
				)

				lspClient.conn = jsonRpcConn

				_, err := lspClient.initialize()
				if err == nil {
					break
				}

				lspClient.didOpen()
			}
		}
	}()

	return &lspClient
}

func waitForConditionWithTimeout(condition func() bool, timeout time.Duration) bool {
	done := make(chan bool)

	go func() {
		for {
			if condition() {
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}

func lspCompletionsToSuggests(completions []lsp.CompletionItem) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	for _, completion := range completions {
		suggestions = append(suggestions, lspCompletionToSuggest(completion))
	}
	return suggestions
}

func lspCompletionToSuggest(completion lsp.CompletionItem) prompt.Suggest {
	return prompt.Suggest{
		Text:        completion.Label,
		Description: completion.Detail,
	}
}
