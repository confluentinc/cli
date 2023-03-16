package autocomplete

import (
	"fmt"
	"github.com/confluentinc/go-prompt"
	"regexp"
	"strings"
	"unicode"
)

func getLastWord(phrase string) string {
	startOfWord := 0
	phrase = strings.TrimSpace(phrase)
	runes := []rune(phrase)
	for i := len(runes) - 1; i >= 0; i-- {
		if unicode.IsSpace(runes[i]) {
			break
		}
		startOfWord = i
	}
	return phrase[startOfWord:]
}

func getNextWord(phrase string) string {
	var word strings.Builder
	for _, char := range strings.TrimSpace(phrase) {
		if unicode.IsSpace(char) {
			break
		}
		word.WriteRune(char)
	}
	return word.String()
}

func isLastCharSpace(str string) bool {
	if len(str) == 0 {
		return false
	}
	runes := []rune(str)
	return unicode.IsSpace(runes[len(runes)-1])
}

// filters all suggestions that start with this prefix and returns only the text after the prefix
func SuggestFromPrefix(suggestions []prompt.Suggest, prefix string) []prompt.Suggest {
	if prefix == "" {
		return suggestions
	}
	//ignore line breaks
	prefix = strings.ReplaceAll(prefix, "\n", " ")
	//ignore case
	prefix = strings.ToUpper(prefix)
	lastWord := getLastWord(prefix)
	lastWord = strings.ToUpper(lastWord)

	isLastWordComplete := isLastCharSpace(prefix)
	ret := make([]prompt.Suggest, 0, len(suggestions))
	for _, suggestion := range suggestions {
		cleanedText := strings.ToUpper(suggestion.Text)
		if strings.HasPrefix(cleanedText, prefix) {
			//only attach a diff of the suggestion and the prefix and include the last word if it was not complete
			text := suggestion.Text[len(prefix):]
			startOfLastWord := strings.LastIndex(prefix, lastWord) //should never be -1 because last word is part of prefix
			if !isLastWordComplete && startOfLastWord != -1 {
				text = suggestion.Text[startOfLastWord:]
			}
			ret = append(ret, prompt.Suggest{
				Text:        text,
				Description: suggestion.Description,
			})
		}
	}
	return ret
}

// filters all suggestions that contain this word and returns only the next word after this word
func SuggestNextWordFromLastWord(suggestions []prompt.Suggest, prefix string) []prompt.Suggest {
	nextWordSuggestions := make([]prompt.Suggest, 0)
	if strings.TrimSpace(prefix) == "" {
		return nextWordSuggestions
	}
	//ignore case
	prefix = strings.ToUpper(prefix)
	lastWord := getLastWord(prefix)
	lastWord = strings.ToUpper(lastWord)

	isLastWordComplete := isLastCharSpace(prefix)
	suggestionSet := map[string]bool{}
	pattern := fmt.Sprintf("\\b%s", regexp.QuoteMeta(lastWord))
	regex, err := regexp.Compile(pattern)
	//avoid crashing the client on regex failure
	if err != nil {
		return nextWordSuggestions
	}
	for _, suggestion := range suggestions {
		posOfLastWord := regex.FindStringIndex(strings.ToUpper(suggestion.Text))
		// if the last word is not present, skip this suggestion
		if posOfLastWord == nil {
			continue
		}

		completeLastWord := ""
		startOfLastWord := posOfLastWord[0]
		startOfNextWord := posOfLastWord[1]
		if !isLastWordComplete {
			completeLastWord = getNextWord(suggestion.Text[startOfLastWord:]) + " "
			startOfNextWord = startOfLastWord + len(completeLastWord)
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
		suggestionKey := replacer.Replace(strings.TrimSpace(completeLastWord + nextWord))
		_, suggestionExists := suggestionSet[suggestionKey]
		suggestionSet[suggestionKey] = true
		if !suggestionExists && suggestionKey != "" {
			nextWordSuggestions = append(nextWordSuggestions, prompt.Suggest{Text: suggestionKey})
		}
	}
	return nextWordSuggestions
}
