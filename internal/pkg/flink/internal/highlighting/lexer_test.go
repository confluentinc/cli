package highlighting

import (
	"strings"
	"testing"

	"github.com/confluentinc/cli/internal/pkg/flink/config"
	"github.com/confluentinc/cli/internal/pkg/flink/test/generators"
	prompt "github.com/confluentinc/go-prompt"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
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
	require.Equal(t, prompt.White, elements[1].Color)
	require.Equal(t, " ", elements[1].Text)
	require.Equal(t, prompt.White, elements[2].Color)
	require.Equal(t, "FIELD", elements[2].Text)
	require.Equal(t, prompt.White, elements[3].Color)
	require.Equal(t, " ", elements[3].Text)
	require.Equal(t, prompt.White, elements[4].Color)
	require.Equal(t, "\n", elements[4].Text)
	require.Equal(t, prompt.Cyan, elements[5].Color)
	require.Equal(t, "FROM", elements[5].Text)
	require.Equal(t, prompt.White, elements[6].Color)
	require.Equal(t, " ", elements[6].Text)
	require.Equal(t, prompt.Cyan, elements[7].Color)
	require.Equal(t, "TABLE", elements[7].Text)
	require.Equal(t, prompt.White, elements[8].Color)
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

		//then
		elements := Lexer(statement)

		for _, element := range elements {
			element.Text = strings.ToUpper(element.Text)
			_, isKeyWord := config.SQLKeywords[element.Text]
			if len(element.Text) == 0 {
				t.Errorf("Empty element in statement: %s", element.Text)
			}

			if isKeyWord {
				require.Equalf(t, element.Color, config.HIGHLIGHT_COLOR, "wrong colour for element: %s", element.Text)
			} else if wrappedInInvertedCommasOrBackticks(element.Text) {
				require.Equalf(t, element.Color, prompt.Yellow, "wrong colour for element: %s", element.Text)
			} else {
				require.Equalf(t, element.Color, prompt.White, "wrong colour for element: %s", element.Text)
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

func TestWordLexerForRandomStatements(t *testing.T) {
	// given
	rapid.Check(t, func(t *rapid.T) {
		randomStatement := generators.RandomSQLSentence().Draw(t, "randomStatement")

		// when
		elements := Lexer(randomStatement.Text)

		for _, element := range elements {
			element.Text = strings.ToUpper(element.Text)
			_, isKeyWord := config.SQLKeywords[element.Text]
			if len(element.Text) == 0 {
				t.Errorf("Empty element in statement: %s", element.Text)
			}

			if isKeyWord {
				require.Equalf(t, element.Color, config.HIGHLIGHT_COLOR, "wrong colour for element: %s", element.Text)
			} else if wrappedInInvertedCommasOrBackticks(element.Text) {
				require.Equalf(t, element.Color, prompt.Yellow, "wrong colour for element: %s", element.Text)
			} else {
				require.Equalf(t, element.Color, prompt.White, "wrong colour for element: %s", element.Text)
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
