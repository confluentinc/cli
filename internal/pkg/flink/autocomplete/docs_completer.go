package autocomplete

import (
	_ "embed"
	"encoding/json"
	"github.com/confluentinc/go-prompt"
	"log"
	"sort"
)

//go:embed code_snippets.json
var codeSnippets []byte
var snippetSuggestions []prompt.Suggest

func loadSnippetSuggestions() {
	var payload map[string]interface{}
	err := json.Unmarshal(codeSnippets, &payload)
	if err != nil {
		log.Printf("Couldn't unmarshal code snippets. Error: %v\n", err)
	}

	for _, value := range payload {
		arr := value.([]interface{})
		for _, example := range arr {
			snippetSuggestions = append(snippetSuggestions, prompt.Suggest{Text: example.(string)})
		}
	}
	//sort result to make order deterministic
	sort.Slice(snippetSuggestions, func(i, j int) bool {
		return snippetSuggestions[i].Text < snippetSuggestions[j].Text
	})
}

func generateDocsCompleter() prompt.Completer {
	loadSnippetSuggestions()
	return docsCompleter
}

func docsCompleter(in prompt.Document) []prompt.Suggest {
	return SuggestNextWordFromLastWord(snippetSuggestions, in.TextBeforeCursor())
}
