package lsp

import (
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/flink-sql-language-service/pkg/api"
	"github.com/confluentinc/go-prompt"
	"github.com/sourcegraph/go-lsp"
)

func LSPCompleter(c LSPInterface, configurationSettings func() api.CliContext) prompt.Completer {
	return func(in prompt.Document) []prompt.Suggest {
		textBeforeCursor := in.TextBeforeCursor()
		if textBeforeCursor == "" {
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
