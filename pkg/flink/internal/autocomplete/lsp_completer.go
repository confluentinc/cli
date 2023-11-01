package autocomplete

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/confluentinc/flink-sql-language-service/pkg/api"
	"github.com/confluentinc/flink-sql-language-service/pkg/server/tcp"
	prompt "github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

type LSPClientInterface interface {
	LSPCompleter(in prompt.Document) []prompt.Suggest
	ShutdownAndExit()
}

type LSPClient struct {
	conn        types.JSONRpcConn
	documentURI *lsp.DocumentURI
	store       types.StoreInterface
}

func (c *LSPClient) LSPCompleter(in prompt.Document) []prompt.Suggest {
	err := c.didChange(in.Text)

	if err != nil {
		log.CliLogger.Debugf("Error sending didChange lsp request: %v\n", err)
		return []prompt.Suggest{}
	}

	textBeforeCursor := in.TextBeforeCursor()

	position := lsp.Position{
		Line:      0,
		Character: len(textBeforeCursor),
	}

	completions := []lsp.CompletionItem{}
	if textBeforeCursor != "" {
		completionList, err := c.completion(position)

		if err != nil {
			log.CliLogger.Debugf("Error sending completion lsp request: %v\n", err)
			return []prompt.Suggest{}
		}

		completions = completionList.Items
	}

	return lspCompletionsToSuggestsOld(completions)
}

func (c *LSPClient) initialize() (*lsp.InitializeResult, error) {
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

func (c *LSPClient) didOpen() error {
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

func (c *LSPClient) didChange(newText string) error {
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

func (c *LSPClient) completion(position lsp.Position) (lsp.CompletionList, error) {
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

func NewLSPClient(store types.StoreInterface) LSPClientInterface {
	lspClient := &LSPClient{
		store: store,
	}

	go func() {
		for port := 49152; port <= 65535; port++ {
			address := fmt.Sprintf("localhost:%d", port)
			lspServer := tcp.NewServer(address)

			go func() {
				err := lspServer.Serve("localhost:8080")
				if err != nil {
					log.CliLogger.Debugf("Coudln't initialize lsp server: %v", err)
				}
			}()

			if waitForConditionWithTimeout(lspServer.IsRunning, 1*time.Second) {
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

					lspInitParams, err := lspClient.initialize()
					log.CliLogger.Trace("LSP init params: ", lspInitParams)

					if err == nil {
						err = lspClient.didOpen()
						if err == nil {
							break
						}
					}
				}
			}
		}
	}()

	return lspClient
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

func lspCompletionsToSuggestsOld(completions []lsp.CompletionItem) []prompt.Suggest {
	suggestions := []prompt.Suggest{}
	for _, completion := range completions {
		suggestions = append(suggestions, lspCompletionToSuggest(completion))
	}
	return suggestions
}

func lspCompletionsToSuggests(completions []lsp.CompletionItem, wordUntilCursor string, startOfPreviousWord int) []prompt.Suggest {

	suggestions := []prompt.Suggest{}
	for _, completion := range completions {
		if completion.TextEdit != nil {
			suggestions = append(suggestions, lspTextEditToSuggestion(completion, wordUntilCursor, startOfPreviousWord))
		} else {
			suggestions = append(suggestions, lspCompletionToSuggest(completion))
		}
	}
	return suggestions
}

func lspCompletionToSuggest(completion lsp.CompletionItem) prompt.Suggest {
	return prompt.Suggest{
		Text:        completion.InsertText,
		Description: completion.Detail,
	}
}

func lspTextEditToSuggestion(completion lsp.CompletionItem, wordUntilCursor string, startOfPreviousWord int) prompt.Suggest {
	replaceRange := completion.TextEdit.Range
	if replaceRange.Start.Line != 0 || replaceRange.End.Character != 0 {
		log.CliLogger.Debug("we only support one statement at the time")
	}

	// We only have to insert text
	if replaceRange.Start.Character == replaceRange.End.Character {
		return prompt.Suggest{
			Text:        wordUntilCursor + completion.TextEdit.NewText,
			Description: completion.Detail,
		}
	} else {
		// we have to replace the text relative to the cursor
		start := replaceRange.Start.Character - startOfPreviousWord
		end := replaceRange.End.Character - startOfPreviousWord

		text := wordUntilCursor[:start] + completion.TextEdit.NewText + wordUntilCursor[end:]
		return prompt.Suggest{
			Text:        text,
			Description: completion.Detail,
		}
	}

}
