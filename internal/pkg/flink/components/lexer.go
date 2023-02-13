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

	for i, word := range words {
		element := prompt.LexerElement{}

		_, isKeyword := sqlKeywords[word]
		if isKeyword {
			element.Color = HIGHLIGHT_COLOR
		} else {
			element.Color = prompt.White
		}

		// We have to maintain the spaces between words if not the last word
		if i != len(words)-1 {
			element.Text = word + "	"
		} else {
			element.Text = word
		}

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}

/* This outputs words all characters in the line with their respective color */
func lexer(line string) []prompt.LexerElement {
	var lexerElements []prompt.LexerElement
	lexerWords := wordLexer(line)

	for _, word := range lexerWords {
		charArr := strings.Split(word.Text, "")

		for _, char := range charArr {
			lexerElements = append(lexerElements, prompt.LexerElement{Color: word.Color, Text: char})
		}
	}

	return lexerElements
}
