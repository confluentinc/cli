package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDoesPathExist(t *testing.T) {
	t.Run("DoesPathExist: empty path returns false", func(t *testing.T) {
		req := require.New(t)
		valid := DoesPathExist("")
		req.False(valid)
	})
}

func TestLoadPropertiesFile(t *testing.T) {
	t.Run("LoadPropertiesFile: empty path yields error", func(t *testing.T) {
		req := require.New(t)
		_, err := LoadPropertiesFile("")
		req.Error(err)
	})
}

func TestAbbreviate(t *testing.T) {
	tests := []struct {
		input    string
		maxLen   int
		expected string
	}{
		{
			input:    "helloooooo",
			maxLen:   3,
			expected: "hel...",
		},
		{
			input:    "helloooooo",
			maxLen:   50,
			expected: "helloooooo",
		},
		{
			input:    "hi",
			maxLen:   2,
			expected: "hi",
		},
	}
	for _, test := range tests {
		out := Abbreviate(test.input, test.maxLen)
		require.Equal(t, test.expected, out)
	}
}

func TestCropString(t *testing.T) {
	for _, test := range []struct {
		s       string
		n       int
		cropped string
	}{
		{"ABCDE", 4, "A..."},
		{"ABCDE", 5, "ABCDE"},
		{"ABCDE", 8, "ABCDE"},
	} {
		require.Equal(t, test.cropped, CropString(test.s, test.n))
	}
}

func TestArrayToCommaDelimitedString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{},
			expected: "",
		},
		{
			input:    []string{"val1"},
			expected: `"val1"`,
		},
		{
			input:    []string{"val1", "val2"},
			expected: `"val1" or "val2"`,
		},
		{
			input:    []string{"val1", "val2", "val3"},
			expected: `"val1", "val2", or "val3"`,
		},
	}
	for _, test := range tests {
		out := ArrayToCommaDelimitedString(test.input, "or")
		require.Equal(t, test.expected, out)
	}
}
