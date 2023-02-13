package components

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"
)

var sqlKeywords = map[string]int{
	"SELECT": 1,
	"FROM":   1,
}

var HIGHLIGHT_COLOR = prompt.Yellow

/* This outputs words with their colors according to if they are flink sql keywords or not */
func wordLexer(line string) []prompt.LexerElement {
	var lexerWords []prompt.LexerElement

	words := strings.Split(line, " ")

	for _, word := range words {
		element := prompt.LexerElement{
			Text: word,
		}

		_, isKeyword := sqlKeywords[word]
		if isKeyword {
			element.Color = HIGHLIGHT_COLOR
		} else {
			element.Color = prompt.White
		}

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}

/* This outputs words all characters in the line with their respective color */
func lexer(line string) []prompt.LexerElement {
	var elements []prompt.LexerElement

	strArr := strings.Split(line, "")

	for k, v := range strArr {
		element := prompt.LexerElement{
			Text: v,
		}

		// every even char must be green.
		if k > 10 && k < 20 {
			element.Color = prompt.Green
		} else {
			element.Color = prompt.White
		}

		elements = append(elements, element)
	}

	return elements
}
