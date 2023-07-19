package autocomplete

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/confluentinc/flink-sql-language-service/pkg/server/tcp"
	prompt "github.com/confluentinc/go-prompt"
	lspInternal "github.com/lighttiger2505/sqls/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

type LSPClientInterface interface {
	LSPCompleter(in prompt.Document) []prompt.Suggest
}

type LSPClient struct {
	conn *jsonrpc2.Conn
}

func (c *LSPClient) LSPCompleter(in prompt.Document) []prompt.Suggest {
	c.didChange(in.Text)
	textBeforeCursor := in.TextBeforeCursor()

	position := lspInternal.Position{
		Line:      0,
		Character: len(textBeforeCursor),
	}

	completions := []lspInternal.CompletionItem{}
	if textBeforeCursor != "" {
		completions = c.completion(position)
	}

	return lspCompletionsToSuggests(completions)
}

func (c *LSPClient) didChange(newText string) {
	var resp interface{}

	didchangeParams := lspInternal.DidChangeTextDocumentParams{
		TextDocument: lspInternal.VersionedTextDocumentIdentifier{
			Version: 2,
			URI:     "test.sql",
		},
		ContentChanges: []lspInternal.TextDocumentContentChangeEvent{
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

func (c *LSPClient) completion(position lspInternal.Position) []lspInternal.CompletionItem {
	var resp []lspInternal.CompletionItem

	completionParams := lspInternal.CompletionParams{TextDocumentPositionParams: lspInternal.TextDocumentPositionParams{
		TextDocument: lspInternal.TextDocumentIdentifier{
			URI: "test.sql",
		},
		Position: position,
	}}

	if c.conn == nil {
		return []lspInternal.CompletionItem{}
	}

	err := c.conn.Call(context.Background(), "textDocument/completion", completionParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	}

	// add proper return type
	return resp
}

func waitForConditionWithTimeout(condition func() bool, timeout time.Duration) bool {
	done := make(chan bool)

	go func() {
		for {
			if condition() {
				done <- true
				return
			}
			time.Sleep(100 * time.Millisecond) // adjust the sleep duration as needed
		}
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
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
				break
			}
		}
	}()

	return &lspClient
}

func lspCompletionsToSuggests(completions []lspInternal.CompletionItem) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	for _, completion := range completions {
		suggestions = append(suggestions, lspCompletionToSuggest(completion))
	}
	return suggestions
}

func lspCompletionToSuggest(completion lspInternal.CompletionItem) prompt.Suggest {
	return prompt.Suggest{
		Text:        completion.Label,
		Description: completion.Detail,
	}
}
