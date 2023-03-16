package components

import (
	"strings"

	"github.com/confluentinc/flink-sql-client/config"
	prompt "github.com/confluentinc/go-prompt"
)

/* This outputs words with their colors according to if they are flink sql keywords or not */
func wordLexer(line string) []prompt.LexerElement {
	lexerWords := []prompt.LexerElement{}

	if line == "" {
		return lexerWords
	}

	words := strings.Split(line, " ")

	for i, word := range words {
		element := prompt.LexerElement{}

		_, isKeyword := config.SQLKeywords[strings.ToUpper(word)]
		if isKeyword {
			element.Color = config.HIGHLIGHT_COLOR
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
func Lexer(line string) []prompt.LexerElement {
	var lexerElements []prompt.LexerElement
	lexerWords := wordLexer(line)

	for _, word := range lexerWords {
		charArr := strings.Split(word.Text, "")

		for _, char := range charArr {
			element := prompt.LexerElement{Color: word.Color, Text: char}

			// Replace empty spaces with white empty string or else the lexer doesn't work properly
			if char == "\t" {
				element.Text = " "
				element.Color = prompt.White
			}

			lexerElements = append(lexerElements, element)
		}
	}

	return lexerElements
}
