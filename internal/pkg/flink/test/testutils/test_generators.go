package testutils

import (
	"strings"

	prompt "github.com/c-bata/go-prompt"

	"github.com/confluentinc/flink-sql-client/config"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"
)

type RandomStatement struct {
	Text          string
	LexerElements []prompt.LexerElement
}

func RandomStatementGenerator(maxWordsCount int32) *rapid.Generator[RandomStatement] {

	sqlKeywordsSlice := maps.Keys(config.SQLKeywords)

	return rapid.Custom(func(t *rapid.T) RandomStatement {
		wordsCount := rapid.Int32Max(maxWordsCount).Draw(t, "Words Count")
		var statement string = ""
		lexerElements := []prompt.LexerElement{}

		for i := int32(0); i < wordsCount; i++ {
			addSqlKeyword := rapid.Bool().Draw(t, "Bool")

			word := rapid.StringN(1, 15, -1).Draw(t, "Word")
			// We need a word, not a phrase and that's why we need to remove empty spaces
			word = strings.TrimSpace(word)
			word = strings.ReplaceAll(word, " ", "")
			word = strings.ToUpper(word)

			// The random generator can generate SQL keywords or strings with unicode
			// characters for empty space (i.e. "\u2003") so we just skip this cases for simplicity
			_, isRandomWordASqlKeyword := config.SQLKeywords[word]
			if isRandomWordASqlKeyword || word == "" {
				i--
				continue
			}

			if addSqlKeyword {
				keywordIndex := rapid.IntRange(0, len(config.SQLKeywords)-1).Draw(t, "Keyword Index")
				keyword := sqlKeywordsSlice[keywordIndex]
				statement += " " + keyword
				lexerElements = append(lexerElements, prompt.LexerElement{Text: keyword, Color: config.HIGHLIGHT_COLOR})
			} else {
				statement += " " + word
				lexerElements = append(lexerElements, prompt.LexerElement{Text: word, Color: prompt.White})
			}
		}

		statement = strings.TrimSpace(statement)
		statement = strings.ToUpper(statement)

		return RandomStatement{statement, lexerElements}
	})
}
