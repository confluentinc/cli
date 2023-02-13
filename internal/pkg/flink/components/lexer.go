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

func TestWordLexer(line string) []prompt.LexerElement {
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

func TestLexer(line string) []prompt.LexerElement {
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
