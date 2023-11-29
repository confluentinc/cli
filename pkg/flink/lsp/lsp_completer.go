package lsp

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/jsonrpc2"

	"github.com/confluentinc/flink-sql-language-service/pkg/server/tcp"
	prompt "github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/flink/types"
	"github.com/confluentinc/cli/v3/pkg/log"
)

type noopHandler struct{}

func (noopHandler) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {}

func NewLocalLSPClient(store types.StoreInterface) LSPInterface {
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

					lspInitParams, err := lspClient.Initialize()
					log.CliLogger.Trace("LSP init params: ", lspInitParams)

					if err == nil {
						err = lspClient.DidOpen()
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
	if replaceRange.Start.Line != 0 {
		log.CliLogger.Debug("we only support replaces with start line 0 for now")
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
