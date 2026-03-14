package formatter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock ErrorWithSuggestions for testing
type mockErrorWithSuggestions struct {
	msg         string
	suggestions string
}

func (e *mockErrorWithSuggestions) Error() string {
	return e.msg
}

func (e *mockErrorWithSuggestions) GetSuggestionsMsg() string {
	return e.suggestions
}

func TestFormatError(t *testing.T) {
	t.Run("error with command context", func(t *testing.T) {
		err := errors.New("resource not found")
		result := FormatError(err, "kafka cluster list", "")

		assert.Contains(t, result, "Command: kafka cluster list")
		assert.Contains(t, result, "Error: resource not found")
	})

	t.Run("error without command context", func(t *testing.T) {
		err := errors.New("authentication failed")
		result := FormatError(err, "", "")

		assert.NotContains(t, result, "Command:")
		assert.Contains(t, result, "Error: authentication failed")
	})

	t.Run("error with suggestions", func(t *testing.T) {
		err := &mockErrorWithSuggestions{
			msg:         "cluster not found",
			suggestions: "List available clusters with `confluent kafka cluster list`.",
		}
		result := FormatError(err, "kafka cluster describe", "")

		assert.Contains(t, result, "Command: kafka cluster describe")
		assert.Contains(t, result, "Error: cluster not found")
		assert.Contains(t, result, "Suggestions:")
		assert.Contains(t, result, "List available clusters with `confluent kafka cluster list`.")
	})

	t.Run("error without suggestions", func(t *testing.T) {
		err := errors.New("simple error")
		result := FormatError(err, "version", "")

		assert.Contains(t, result, "Error: simple error")
		assert.NotContains(t, result, "Suggestions:")
	})

	t.Run("rawOutput passthrough when starts with Error:", func(t *testing.T) {
		err := errors.New("wrapped error")
		rawOutput := "Error: original error message\ndetails here"
		result := FormatError(err, "test command", rawOutput)

		assert.Equal(t, rawOutput, result)
		assert.NotContains(t, result, "wrapped error")
	})

	t.Run("no symbols or emoji in output", func(t *testing.T) {
		err := errors.New("test error")
		result := FormatError(err, "test", "")

		// Ensure no symbols/emoji
		assert.NotContains(t, result, "✓")
		assert.NotContains(t, result, "✗")
		assert.NotContains(t, result, "⚠")
		assert.NotContains(t, result, "❌")
		assert.NotContains(t, result, "✅")
	})

	t.Run("multiline suggestions formatted correctly", func(t *testing.T) {
		err := &mockErrorWithSuggestions{
			msg:         "invalid configuration",
			suggestions: "Check your API key with `confluent api-key list`.\nVerify your environment with `confluent environment list`.",
		}
		result := FormatError(err, "login", "")

		assert.Contains(t, result, "Suggestions:")
		assert.Contains(t, result, "Check your API key")
		assert.Contains(t, result, "Verify your environment")
	})
}
