package reverseisearch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
)

func TestSearchString(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// create a random array string
		slice := rapid.SliceOfN(rapid.StringN(1, -1, -1), 1, 500).Draw(t, "Slice of strings")
		// randomly pick one of the string inside the array
		index := rapid.IntRange(0, len(slice)-1).Draw(t, "Index")
		s := slice[index]

		result := search(s, slice, len(slice)-1)
		assert.NotEqual(t, -1, result.index)
		assert.Contains(t, result.match, s)
	})
}

func TestSearchWithIndex(t *testing.T) {
	slice := []string{"first", "second", "third one", "third two", "third three"}

	// last element
	result := search("third", slice, len(slice)-1)

	assert.Equal(t, len(slice)-1, result.index)
	assert.Equal(t, "third three", result.match)

	// last element -1
	result = search("third", slice, len(slice)-2)

	assert.Equal(t, len(slice)-2, result.index)
	assert.Equal(t, "third two", result.match)

	// last element - 2
	result = search("third", slice, len(slice)-3)

	assert.Equal(t, len(slice)-3, result.index)
	assert.Equal(t, "third one", result.match)

	// last element
	result = search("third", slice, 0)

	assert.Equal(t, -1, result.index)
	assert.Equal(t, "", result.match)
}

func TestSearchWithOutOfBoundIndex(t *testing.T) {
	slice := []string{"first", "second", "third one", "third two", "third three"}

	// when out of bound index, will just start from last element
	result := search("third", slice, len(slice)+100)
	assert.Equal(t, 4, result.index)
	assert.Equal(t, "third three", result.match)

	// negative indexes are just ignored
	result = search("third", slice, -100)

	assert.Equal(t, -1, result.index)
	assert.Equal(t, "", result.match)
}

func TestNoMatchEmptyString(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// create a random array string
		slice := rapid.SliceOfN(rapid.StringN(1, -1, -1), 1, 500).Draw(t, "Slice of strings")

		res := search("", slice, len(slice)-1)
		assert.Equal(t, -1, res.index)
		assert.Equal(t, "", res.match)
	})
}

func TestNoMatchString(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		slice := rapid.SliceOfN(rapid.StringMatching("[0-9]"), 1, 500).Draw(t, "Slice of strings")
		s := rapid.StringMatching("[a-zA-Z]").Draw(t, "Literal String")

		res := search(s, slice, len(slice)-1)
		assert.Equal(t, -1, res.index)
		assert.Equal(t, "", res.match)
	})
}

func TestNewLineCount(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		newLines := rapid.IntRange(0, 100).Draw(t, "Number of newLine")
		s := rapid.StringMatching("\\w{4,}").Draw(t, "Random String")

		for i := 0; i < newLines; i++ {
			randomIndex := rapid.IntRange(0, len(s)-2).Draw(t, "Index")
			s = s[randomIndex:] + "\n" + s[:randomIndex]
		}

		assert.Equal(t, newLines, NewLineCount(s))
	})
}
