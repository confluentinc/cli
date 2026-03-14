package formatter

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFormatter(t *testing.T) {
	f := NewFormatter()
	assert.NotNil(t, f)
}

func TestFormat_StripANSIFirst(t *testing.T) {
	f := NewFormatter()

	rawOutput := "\x1b[32mSuccess\x1b[0m: Operation completed"
	result := f.Format(rawOutput, nil, "test command")

	require.NotNil(t, result)
	// Raw field should preserve original output
	assert.Equal(t, rawOutput, result.Raw)
	// Summary should have ANSI codes stripped
	assert.NotContains(t, result.Summary, "\x1b[32m")
	assert.NotContains(t, result.Summary, "\x1b[0m")
	assert.Contains(t, result.Summary, "Success")
}

func TestFormat_PlainText(t *testing.T) {
	f := NewFormatter()

	rawOutput := "Plain text output"
	result := f.Format(rawOutput, nil, "test command")

	require.NotNil(t, result)
	assert.Equal(t, rawOutput, result.Raw)
	assert.Equal(t, "Plain text output", result.Summary)
	assert.Nil(t, result.Structured)
}

func TestFormat_EmptyString(t *testing.T) {
	f := NewFormatter()

	result := f.Format("", nil, "test command")

	require.NotNil(t, result)
	assert.Equal(t, "", result.Raw)
	assert.Equal(t, "", result.Summary)
	assert.Nil(t, result.Structured)
}

func TestFormat_ErrorPath(t *testing.T) {
	f := NewFormatter()

	rawOutput := "Command failed"
	execErr := errors.New("execution error")
	result := f.Format(rawOutput, execErr, "test command")

	require.NotNil(t, result)
	assert.Equal(t, rawOutput, result.Raw)
	// Error formatter should be called (will be implemented properly in Task 2)
	assert.NotEmpty(t, result.Summary)
	assert.Contains(t, result.Summary, "error")
}

func TestFormat_JSONArray(t *testing.T) {
	f := NewFormatter()

	jsonOutput := `[{"id":"lkc-1","name":"prod"},{"id":"lkc-2","name":"dev"}]`
	result := f.Format(jsonOutput, nil, "kafka cluster list")

	require.NotNil(t, result)
	assert.Equal(t, jsonOutput, result.Raw)
	// Summary should contain formatted list (Task 2 will implement fully)
	assert.NotEmpty(t, result.Summary)
	// Structured should contain parsed JSON
	assert.NotNil(t, result.Structured)
}

func TestFormat_JSONObject(t *testing.T) {
	f := NewFormatter()

	jsonOutput := `{"id":"lkc-123","name":"prod","status":"UP"}`
	result := f.Format(jsonOutput, nil, "kafka cluster describe")

	require.NotNil(t, result)
	assert.Equal(t, jsonOutput, result.Raw)
	// Summary should contain formatted object (Task 2 will implement fully)
	assert.NotEmpty(t, result.Summary)
	// Structured should contain parsed JSON
	assert.NotNil(t, result.Structured)
}

func TestFormattedOutput_Structure(t *testing.T) {
	// Test that FormattedOutput struct has required fields
	output := &FormattedOutput{
		Summary:    "test summary",
		Structured: map[string]interface{}{"key": "value"},
		Raw:        "raw output",
	}

	assert.Equal(t, "test summary", output.Summary)
	assert.Equal(t, "raw output", output.Raw)
	assert.NotNil(t, output.Structured)

	structured, ok := output.Structured.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "value", structured["key"])
}

func TestFormat_ANSIWithJSON(t *testing.T) {
	f := NewFormatter()

	// Realistic case: JSON output with ANSI codes from lipgloss
	rawOutput := "\x1b[32m[{\"id\":\"lkc-1\",\"name\":\"prod\"}]\x1b[0m"
	result := f.Format(rawOutput, nil, "kafka cluster list")

	require.NotNil(t, result)
	// Raw preserves ANSI
	assert.Contains(t, result.Raw, "\x1b[32m")
	// Summary should have ANSI stripped before JSON processing
	assert.NotContains(t, result.Summary, "\x1b[32m")
	assert.NotContains(t, result.Summary, "\x1b[0m")
}
