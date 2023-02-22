package autocomplete

import (
	_ "embed"
	"encoding/json"
	"github.com/c-bata/go-prompt"
	"log"
	"strings"
	"unicode"
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
}

func generateDocsCompleter() prompt.Completer {
	loadSnippetSuggestions()
	return docsCompleter
}

func docsCompleter(in prompt.Document) []prompt.Suggest {
	lastWord := in.GetWordBeforeCursorWithSpace()
	return suggestNextWordFromLastWord(snippetSuggestions, lastWord)
}

func suggestNextWordFromLastWord(suggestions []prompt.Suggest, lastWord string) []prompt.Suggest {
	nextWordSuggestions := make([]prompt.Suggest, 0)
	if strings.TrimSpace(lastWord) == "" {
		return nextWordSuggestions
	}

	suggestionSet := map[string]bool{}
	isLastWordComplete := isLastCharSpace(lastWord)
	lastWord = " " + lastWord //only look for whole word
	for _, suggestion := range suggestions {
		idxOfLastWord := strings.Index(strings.ToLower(suggestion.Text), strings.ToLower(lastWord))
		// if the last word is not present, skip this suggestion
		if idxOfLastWord == -1 {
			continue
		}

		completeLastWord := ""
		startOfNextWord := idxOfLastWord + len(lastWord)
		if !isLastWordComplete {
			completeLastWord = getNextWord(suggestion.Text[idxOfLastWord:]) + " "
			startOfNextWord = idxOfLastWord + len(completeLastWord)
			//make sure to not step out of bounds if this is the last word
			if startOfNextWord > len(suggestion.Text) {
				startOfNextWord = len(suggestion.Text)
			}
		}
		nextWord := getNextWord(suggestion.Text[startOfNextWord:])
		/*
			TODO:
			Currently we are replacing special chars like parentheses from the next word, which could potentially
			make the suggestion syntactically incorrect. It would probably be better to have a more sophisticated logic here.
			E.g. a method that takes in a suggestion and returns if this is a syntactically correct suggestion.
		*/
		replacer := strings.NewReplacer(
			")", "",
			"(", "",
		)
		//filter out duplicated suggestions
		suggestionKey := strings.TrimSpace(completeLastWord + replacer.Replace(nextWord))
		_, suggestionExists := suggestionSet[suggestionKey]
		suggestionSet[suggestionKey] = true
		if !suggestionExists {
			nextWordSuggestions = append(nextWordSuggestions, prompt.Suggest{Text: suggestionKey})
		}
	}
	return nextWordSuggestions
}

func getNextWord(phrase string) string {
	var word strings.Builder
	for _, char := range phrase {
		// if we find a space we should either skip it because it's a leading white space
		// or we should break, because it marks the end of the word
		if unicode.IsSpace(char) {
			if word.Len() == 0 {
				continue
			}
			break
		}

		word.WriteRune(char)
	}
	return word.String()
}

func isLastCharSpace(word string) bool {
	if len(word) == 0 {
		return false
	}
	return word[len(word)-1:] == " "
}
