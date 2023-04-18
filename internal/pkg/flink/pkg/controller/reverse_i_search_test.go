package controller

import (
	"github.com/stretchr/testify/assert"
	"pgregory.net/rapid"
	"testing"
)

func TestSearchMatchString(t *testing.T) {

	rapid.Check(t, func(t *rapid.T) {
		// create a random array string
		slice := rapid.SliceOfN(rapid.StringN(1, -1, -1), 1, 500).Draw(t, "Slice of strings")
		// randomly pick one of the string inside the array
		index := rapid.IntRange(0, len(slice)-1).Draw(t, "Index")
		s := slice[index]

		result, i := search(s, slice)
		assert.NotEqual(t, -1, i)
		assert.Contains(t, result, s)

	})
}

func TestNoMatchEmptyString(t *testing.T) {

	rapid.Check(t, func(t *rapid.T) {
		// create a random array string
		slice := rapid.SliceOfN(rapid.StringN(1, -1, -1), 1, 500).Draw(t, "Slice of strings")

		res, i := search("", slice)
		assert.Equal(t, -1, i)
		assert.Equal(t, "", res)

	})
}

func TestNoMatchString(t *testing.T) {

	rapid.Check(t, func(t *rapid.T) {

		slice := rapid.SliceOfN(rapid.StringMatching("[0-9]"), 1, 500).Draw(t, "Slice of strings")
		s := rapid.StringMatching("[a-zA-Z]").Draw(t, "Literal String")

		res, i := search(s, slice)
		assert.Equal(t, -1, i)
		assert.Equal(t, "", res)

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
