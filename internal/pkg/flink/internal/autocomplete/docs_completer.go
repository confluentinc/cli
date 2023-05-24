package autocomplete

import (
	_ "embed"
	"encoding/json"
	"log"
	"sort"

	"github.com/confluentinc/go-prompt"
)

//go:embed code_snippets.json
var codeSnippets []byte

func loadSnippetSuggestions(snippetSuggestions []prompt.Suggest) {
	var payload map[string]any
	err := json.Unmarshal(codeSnippets, &payload)
	if err != nil {
		log.Printf("Couldn't unmarshal code snippets. Error: %v\n", err)
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
}

func GenerateDocsCompleter() prompt.Completer {
	var snippetSuggestions []prompt.Suggest
	loadSnippetSuggestions(snippetSuggestions)
	return docsCompleter(snippetSuggestions)
}

func docsCompleter(snippetSuggestions []prompt.Suggest) prompt.Completer {
	return func(in prompt.Document) []prompt.Suggest {
		return SuggestNextWord(snippetSuggestions, in.TextBeforeCursor())
	}
}
