package history

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	pconfig "github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	pversion "github.com/confluentinc/cli/v4/pkg/version"
)

func TestLoadHistory(t *testing.T) {
	history := LoadHistory()

	require.NotNil(t, history, "Expected non-nil history object")
	require.NotEmpty(t, history.confluentPath, "Expected non-empty confluent path")
	require.NotEmpty(t, history.historyPath, "Expected non-empty history path")

	// Create temporary file
	tmpDir, err := os.MkdirTemp("", "confluent-test")
	defer os.RemoveAll(tmpDir)

	require.NoError(t, err, "Error creating temp dir")

	tmpFile := filepath.Join(tmpDir, filename)

	// Write sample data
	sampleData := `{"data":["statement1;","statement2"]}`
	err = os.WriteFile(tmpFile, []byte(sampleData), 0644)

	require.NoError(t, err, "Error writing sample data to temp file")

	// Update historical data path
	prevPath := history.historyPath
	history.historyPath = tmpFile

	// Reload history
	history = loadFromPath(history)
	require.NotNil(t, history, "Expected non-nil history object after reloading history")
	require.NotEmpty(t, history.confluentPath, "Expected non-empty confluent path after reloading history")
	require.NotEmpty(t, history.historyPath, "Expected non-empty history path after reloading history")

	// Verify data
	require.Len(t, history.Data, 2, "Expected two history items")
	require.Equal(t, "statement1;", history.Data[0], "Expected statement1; as first item in history")

	// Reset history path
	history.historyPath = prevPath
}

func TestInitPath_ScopesByChannel(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	t.Setenv(config.HomeConfluentPathEnvVar, "")
	t.Cleanup(func() { pversion.SetProcessChannel(pversion.Dev) })

	tests := []struct {
		name    string
		channel pversion.Channel
		dir     string
	}{
		{"stable keeps the historical path", pversion.Stable, ".confluent"},
		{"dev is isolated", pversion.Dev, ".confluent-dev"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pversion.SetProcessChannel(test.channel)

			history := initPath(filename)

			require.NotNil(t, history)
			require.Equal(t, filepath.Join(home, test.dir), history.confluentPath)
		})
	}
}

func TestInitPath_HomeConfluentPathOverridesChannel(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)
	t.Setenv(config.HomeConfluentPathEnvVar, ".confluent-custom")
	t.Cleanup(func() { pversion.SetProcessChannel(pversion.Dev) })
	pversion.SetProcessChannel(pversion.Stable)

	history := initPath(filename)

	require.NotNil(t, history)
	require.Equal(t, filepath.Join(home, ".confluent-custom"), history.confluentPath,
		"HOME_CONFLUENT_PATH must win over the channel default")
	require.NotEqual(t, filepath.Join(home, pconfig.StateDirName()), history.confluentPath,
		"the override must actually diverge from the channel default it is overriding")
}

func TestHistorySave(t *testing.T) {
	// Create a temp directory
	tmpDir, _ := os.MkdirTemp("", "confluent-test")
	defer os.RemoveAll(tmpDir)

	history := &History{
		confluentPath: tmpDir,
		historyPath:   filepath.Join(tmpDir, filename),
		Data:          []string{"statement1", "statement2"},
	}

	// Save the history
	history.Save()

	// Check if file exists and has the correct content
	fileContent, err := os.ReadFile(history.historyPath)
	require.NoError(t, err)
	expectedJSON := `{"data":["statement1","statement2"]}`
	require.Equal(t, expectedJSON, string(fileContent))

	// Check that the history was truncated to 500 entries
	history.Data = append(history.Data, make([]string, 501)...)
	history.Save()
	require.Len(t, history.Data, 500)
}

func TestAppendHistory(t *testing.T) {
	// Create a History instance for testing
	history := &History{
		Data: []string{"statement1"},
	}

	// Append statements to history with correct format
	history.Append("statement2", "statement3")
	expectedData := []string{"statement1", "statement2", "statement3"}
	require.Equal(t, expectedData, history.Data)

	// Append empty list of statements, should not modify history
	history.Append()
	require.Equal(t, expectedData, history.Data)

	// Append statements without trimming white spaces in front and back
	history.Append(" statement4 ", "\tstatement5\t")
	expectedData = []string{"statement1", "statement2", "statement3", " statement4 ", "\tstatement5\t"}
	require.Equal(t, expectedData, history.Data)
}

func TestFormatStatements(t *testing.T) {
	cases := []struct {
		statements     []string
		expectedOutput []string
	}{
		{[]string{"SELECT * FROM orders;", "SELECT * FROM products;"}, []string{"SELECT * FROM orders;", "SELECT * FROM products;"}},
		{[]string{"INSERT INTO users (name, age) VALUES ('Alice', 25);"}, []string{"INSERT INTO users (name, age) VALUES ('Alice', 25);"}},
		{[]string{"", "SELECT * FROM users;", ""}, []string{"SELECT * FROM users;"}},
		{[]string{"SELECT * FROM users;", "SELECT * FROM users;"}, []string{"SELECT * FROM users;"}},
		{[]string{}, []string{}},
	}

	for _, c := range cases {
		actualOutput := formatStatements(c.statements)
		require.Equal(t, c.expectedOutput, actualOutput)
	}
}
