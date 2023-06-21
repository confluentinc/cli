package autocomplete

import (
	"context"
	"fmt"
	prompt "github.com/confluentinc/go-prompt"
	lspInternal "github.com/lighttiger2505/sqls/pkg/lsp"
)

//var version int = 2

func LSPCompleter(in prompt.Document) []prompt.Suggest {
	didChange(in.Text)
	textBeforeCursor := in.TextBeforeCursor()

	position := lspInternal.Position{
		Line:      0,
		Character: len(textBeforeCursor),
	}

	completions := []lspInternal.CompletionItem{}
	if textBeforeCursor != "" {
		completions = completion(position)
	}

	return lspCompletionsToSuggests(completions)
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

func didChange(newText string) {
	var resp interface{}
	//version++

	didchangeParams := lspInternal.DidChangeTextDocumentParams{
		TextDocument: lspInternal.VersionedTextDocumentIdentifier{
			Version: 2,
			URI:     "test.sql",
		},
		ContentChanges: []lspInternal.TextDocumentContentChangeEvent{
			{Text: newText},
		},
	}

	err := LSP.Call(context.Background(), "textDocument/didChange", didchangeParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	} else {
		fmt.Println("response didChange: ", resp)
	}
}

func completion(position lspInternal.Position) []lspInternal.CompletionItem {
	var resp []lspInternal.CompletionItem

	completionParams := lspInternal.CompletionParams{TextDocumentPositionParams: lspInternal.TextDocumentPositionParams{
		TextDocument: lspInternal.TextDocumentIdentifier{
			URI: "test.sql",
		},
		Position: position,
	}}

	err := LSP.Call(context.Background(), "textDocument/completion", completionParams, &resp)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
	} else {
		fmt.Println("response completion: ", resp)
	}

	// add proper return type
	return resp
}
