package autocomplete

import (
	_ "embed"
	"encoding/json"
	"sort"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/internal/pkg/log"
)

//go:embed code_snippets.json
var codeSnippets []byte

func loadSnippetSuggestions() []prompt.Suggest {
	var snippetSuggestions []prompt.Suggest
	var payload map[string]any
	if err := json.Unmarshal(codeSnippets, &payload); err != nil {
		log.CliLogger.Warnf("Couldn't unmarshal code snippets. Error: %v\n", err)
	}

	for _, value := range payload {
		arr := value.([]any)
		for _, example := range arr {
			snippetSuggestions = append(snippetSuggestions, prompt.Suggest{Text: example.(string)})
		}
	}
	// sort result to make order deterministic
	sort.Slice(snippetSuggestions, func(i, j int) bool {
		return snippetSuggestions[i].Text < snippetSuggestions[j].Text
	})
	return snippetSuggestions
}

func GenerateDocsCompleter() prompt.Completer {
	snippetSuggestions := loadSnippetSuggestions()
	return docsCompleter(snippetSuggestions)
}

func docsCompleter(snippetSuggestions []prompt.Suggest) prompt.Completer {
	return func(in prompt.Document) []prompt.Suggest {
		return SuggestNextWord(snippetSuggestions, in.TextBeforeCursor())
	}
}
