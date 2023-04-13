package lexer

import (
	"strings"

	"github.com/confluentinc/flink-sql-client/config"
	prompt "github.com/confluentinc/go-prompt"
)

type Word struct {
	Text      string
	Separator string
}

var SpecialSplitTokens = map[int32]uint8{'\t': 1, '\n': 1, '\v': 1, '\f': 1, '\r': 1, ' ': 1, ';': 1, '=': 1, '<': 1, '>': 1}

func splitWithSeparators(line string) []string {
	words := []string{}
	word := ""

	for _, char := range line {
		if _, ok := SpecialSplitTokens[char]; ok {
			if word != "" {
				words = append(words, word)
			}
			words = append(words, string(char))
			word = ""
		} else {
			word += string(char)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

func isInvertedCommasWord(word string) bool {
	return word[0] == '\'' && word[len(word)-1] == '\''
}

/* This outputs words all characters in the line with their respective color */
func Lexer(line string) []prompt.LexerElement {
	lexerWords := []prompt.LexerElement{}

	if line == "" {
		return lexerWords
	}

	words := splitWithSeparators(line)

	for _, word := range words {
		element := prompt.LexerElement{}

		_, isKeyword := config.SQLKeywords[strings.ToUpper(strings.TrimSpace(word))]
		if isKeyword {
			element.Color = config.HIGHLIGHT_COLOR
		} else if isInvertedCommasWord(word) {
			element.Color = prompt.Yellow
		} else {
			element.Color = prompt.White
		}

		// We have to maintain the spaces between words if not the last word
		element.Text = word

		lexerWords = append(lexerWords, element)
	}

	return lexerWords
}
