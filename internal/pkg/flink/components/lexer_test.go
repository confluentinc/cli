package components

import (
	"testing"

	prompt "github.com/c-bata/go-prompt"
)

// property based testing?
func TestLexer(t *testing.T) {
	// given
	line := "SELECT FIELD FROM TABLE;"

	// when
	elements := lexer(line)

	// then
	for i, element := range elements {
		if i > 0 && i < 6 {
			if element.Color != HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, HIGHLIGHT_COLOR)
			}
		}
	}
}

func TestWordLexer(t *testing.T) {
	// given
	line := "SELECT FIELD FROM TABLE WHERE FIELD = 2;"

	// when
	elements := wordLexer(line)

	// then
	for _, element := range elements {

		_, isKeyWord := sqlKeywords[element.Text]

		if isKeyWord && element.Color != HIGHLIGHT_COLOR {
			t.Errorf("lexer() = %q, want %q", element.Color, HIGHLIGHT_COLOR)
		} else if !isKeyWord && element.Color != prompt.White {
			t.Errorf("lexer() = %q, want %q", element.Color, HIGHLIGHT_COLOR)
		}

	}
}
