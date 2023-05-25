package generators

import (
	"strings"

	"github.com/samber/lo"
	"golang.org/x/exp/maps"
	"pgregory.net/rapid"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
)

type RandomStatement struct {
	Text string
}

type SQLSentence struct {
	Text       string
	TokenCount int
	Tokens     []string
}

func RandomSQLSentence() *rapid.Generator[SQLSentence] {
	return rapid.Custom(func(t *rapid.T) SQLSentence {
		words := rapid.SliceOfNDistinct(
			rapid.SampledFrom(config.SQLKeywords.Slice()),
			1,
			10,
			rapid.ID[string],
		).Draw(t, "words")
		regularWords := rapid.SliceOfNDistinct(
			rapid.StringMatching("[a-zA-Z]+"),
			1,
			10,
			rapid.ID[string],
		).Draw(t, "regular words")

		words = append(words, regularWords...)
		shuffled := FisherYatesShuffle(words).Draw(t, "shuffled")
		splitTokens := rapid.SliceOfNDistinct(
			rapid.SampledFrom(maps.Keys(config.SpecialSplitTokens)),
			len(shuffled)+1,
			len(shuffled)+1,
			rapid.ID[int32],
		).Draw(t, "split tokens")

		stopwords := lo.Map(splitTokens, func(token int32, _ int) string { return string(token) })
		shuffled = lo.Interleave(shuffled, stopwords)

		line := strings.Join(shuffled, "")

		return SQLSentence{Text: line, TokenCount: len(shuffled), Tokens: shuffled}
	})
}

func FisherYatesShuffle[T any](src []T) *rapid.Generator[[]T] {
	return rapid.Custom(func(t *rapid.T) []T {
		slice := append([]T(nil), src...)

		for i := len(slice) - 1; i > 0; i-- {
			j := rapid.IntRange(0, i).Draw(t, "j")
			slice[i], slice[j] = slice[j], slice[i]
		}

		return slice
	})
}
