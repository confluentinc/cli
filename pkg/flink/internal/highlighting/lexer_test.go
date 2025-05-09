package highlighting

import (
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v4/pkg/color"
	"github.com/confluentinc/cli/v4/pkg/flink/config"
	"github.com/confluentinc/cli/v4/pkg/flink/test/generators"
)

func TestLexer(t *testing.T) {
	// given
	line := "SELECT FIELD \nFROM TABLE;"

	// when
	elements := Lexer(line)

	// then
	require.Len(t, elements, 9)
	require.Equal(t, prompt.Cyan, elements[0].Color)
	require.Equal(t, "SELECT", elements[0].Text)
	require.Equal(t, prompt.DefaultColor, elements[1].Color)
	require.Equal(t, " ", elements[1].Text)
	require.Equal(t, prompt.DefaultColor, elements[2].Color)
	require.Equal(t, "FIELD", elements[2].Text)
	require.Equal(t, prompt.DefaultColor, elements[3].Color)
	require.Equal(t, " ", elements[3].Text)
	require.Equal(t, prompt.DefaultColor, elements[4].Color)
	require.Equal(t, "\n", elements[4].Text)
	require.Equal(t, prompt.Cyan, elements[5].Color)
	require.Equal(t, "FROM", elements[5].Text)
	require.Equal(t, prompt.DefaultColor, elements[6].Color)
	require.Equal(t, " ", elements[6].Text)
	require.Equal(t, prompt.Cyan, elements[7].Color)
	require.Equal(t, "TABLE", elements[7].Text)
	require.Equal(t, prompt.DefaultColor, elements[8].Color)
	require.Equal(t, ";", elements[8].Text)
}

func TestIsLexerCaseInsensitive(t *testing.T) {
	// given
	line := "select field from table;"
	// when
	elements := Lexer(line)

	uppercase := Lexer(strings.ToUpper(line))
	// then
	require.Equal(t, len(elements), len(uppercase))

	for i, element := range elements {
		require.Equal(t, element.Color, uppercase[i].Color)
		require.Equal(t, strings.ToUpper(element.Text), uppercase[i].Text)
	}
}

func TestExamplesWordLexer(t *testing.T) {
	// given
	statements := []string{"SELECT FIELD FROM TABLE WHERE FIELD = 2;",
		"SELECT 'Hello World';",
		"DROP TABLE Orders;",
		"ALTER TABLE Orders RENAME TO NewOrders;",
		"INSERT INTO Orders VALUES ('pen', 2);",
		"ANALYZE TABLE Orders COMPUTE STATISTICS;",
		"DESCRIBE Orders;",
		"EXPLAIN PLAN FOR ...;",
		"USE db1;",
		"CREATE TABLE employee_information (    emp_id INT,    name VARCHAR,    dept_id INT) WITH (     'connector' = 'filesystem',    'path' = '/path/to/something.csv',    'format' = 'csv');",
		"SHOW TABLES;",
		"SET 'table.local-time-zone;' = 'Europe/Berlin';",
		"SELECT * from employee_information WHERE dept_id = 1;",
		"SELECT   dept_id,   COUNT(*) as emp_count FROM employee_information GROUP BY dept_id;",
	}

	// when
	for _, statement := range statements {
		// then
		elements := Lexer(statement)

		for _, element := range elements {
			if element.Text == "" {
				t.Error("empty element in statement")
			}

			if config.SQLKeywords.Contains(strings.ToUpper(element.Text)) {
				require.Equalf(t, element.Color, color.PromptAccentColor, "wrong colour for element: %s", element.Text)
			} else if wrappedInInvertedCommasOrBackticks(element.Text) {
				require.Equalf(t, element.Color, prompt.Yellow, "wrong colour for element: %s", element.Text)
			} else {
				require.Equalf(t, element.Color, prompt.DefaultColor, "wrong colour for element: %s", element.Text)
			}
		}
	}
}

func TestSplitWithSeparators(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// given
		sentence := generators.RandomSQLSentence().Draw(t, "line")
		// when
		tokens := splitWithSeparators(sentence.Text)
		// then
		require.Equal(t, sentence.TokenCount, len(tokens))
	})
}

func TestSplitWithSeparatorsDoesNotIncludeEmptyString(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		// given
		sentence := generators.RandomSQLSentence().Draw(t, "line")
		// when
		tokens := splitWithSeparators(sentence.Text)
		// then
		require.NotContains(t, tokens, "")
	})
}

func TestSplitWithSeparatorsSnapshot(t *testing.T) {
	// given
	sentence := `SELECT count(col1) FROM users \n /*\t\v\f\r[testing]*/WHERE (name = 'John Doe'); -- \n>.,<:= testing)`
	// when
	tokens := splitWithSeparators(sentence)
	// then
	cupaloy.SnapshotT(t, tokens)
}

func TestWordLexerForRandomStatements(t *testing.T) {
	// given
	rapid.Check(t, func(t *rapid.T) {
		randomStatement := generators.RandomSQLSentence().Draw(t, "randomStatement")

		// when
		elements := Lexer(randomStatement.Text)

		for _, element := range elements {
			if element.Text == "" {
				t.Error("empty element in statement")
			}

			if config.SQLKeywords.Contains(strings.ToUpper(element.Text)) {
				require.Equalf(t, element.Color, color.PromptAccentColor, "wrong colour for element: %s", element.Text)
			} else if wrappedInInvertedCommasOrBackticks(element.Text) {
				require.Equalf(t, element.Color, prompt.Yellow, "wrong colour for element: %s", element.Text)
			} else {
				require.Equalf(t, element.Color, prompt.DefaultColor, "wrong colour for element: %s", element.Text)
			}
		}
	})
}

func TestIsInvertedCommasWord(t *testing.T) {
	tables := []struct {
		word     string
		expected bool
	}{
		{"'hello'", true},
		{"`world`", true},
		{"hello", false},
		{"'hello", false},
		{"hello'", false},
	}

	for _, table := range tables {
		result := wrappedInInvertedCommasOrBackticks(table.word)
		if result != table.expected {
			t.Errorf("Expected isInvertedCommasWord(%v) to be %v, but got %v", table.word, table.expected, result)
		}
	}
}
