package lsp

import (
	"github.com/sourcegraph/go-lsp"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/log"
)

func LSPCompleter(c LSPInterface, configurationSettings func() CliContext) prompt.Completer {
	return func(in prompt.Document) []prompt.Suggest {
		textBeforeCursor := in.TextBeforeCursor()
		if textBeforeCursor == "" || c == nil {
			return nil
		}

		err := c.DidChange(in.Text)
		if err != nil {
			log.CliLogger.Debugf("Error sending didChange lsp notification: %v\n", err)
			return nil
		}

		err = c.DidChangeConfiguration(configurationSettings())
		if err != nil {
			log.CliLogger.Debugf("Error sending didChangeConfiguration lsp notification: %v\n", err)
			return nil
		}

		position := lsp.Position{
			Line:      0,
			Character: len(textBeforeCursor),
		}

		completionList, err := c.Completion(position)
		if err != nil {
			log.CliLogger.Debugf("Error sending completion lsp request: %v\n", err)
			return nil
		}

		return lspCompletionsToSuggests(completionList.Items, in.GetWordBeforeCursor(), in.FindStartOfPreviousWord())
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
