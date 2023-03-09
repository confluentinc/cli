package components

import (
	"strings"
	"testing"

	prompt "github.com/c-bata/go-prompt"
	"github.com/confluentinc/flink-sql-client/config"
	testutils "github.com/confluentinc/flink-sql-client/test/testutils"
	"github.com/stretchr/testify/require"
	"pgregory.net/rapid"
)

func TestBasicLexer(t *testing.T) {
	// given
	line := "SELECT FIELD FROM TABLE;"

	// when
	elements := Lexer(line)

	// then
	for i, element := range elements {
		if i >= 0 && i < 6 || i > 12 && i < 17 {
			if element.Color != config.HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, config.HIGHLIGHT_COLOR)
			}
		} else if element.Color != prompt.White {
			t.Errorf("lexer() = %d, want %d", element.Color, prompt.White)
		}

	}
}

func TestIsLexerCaseInsensitive(t *testing.T) {
	// given
	line := "select field from table;"

	// when
	elements := Lexer(line)

	// then
	for i, element := range elements {
		if i >= 0 && i < 6 || i > 12 && i < 17 {
			if element.Color != config.HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, config.HIGHLIGHT_COLOR)
			}
		} else if element.Color != prompt.White {
			t.Errorf("lexer() = %d, want %d", element.Color, prompt.White)
		}

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
		elements := wordLexer(statement)

		for _, element := range elements {
			element.Text = strings.TrimSpace(element.Text)
			element.Text = strings.ToUpper(element.Text)
			_, isKeyWord := config.SQLKeywords[element.Text]

			if isKeyWord {
				require.Equal(t, element.Color, config.HIGHLIGHT_COLOR)
			} else {
				require.Equal(t, element.Color, prompt.White)
			}
		}

	}
}

func TestWordLexerForRandomStatements(t *testing.T) {
	// given
	statementGenerator := testutils.RandomStatementGenerator(15)
	rapid.Check(t, func(t *rapid.T) {
		randomStatement := statementGenerator.Example()

		// when
		realElements := wordLexer(randomStatement.Text)

		// then
		require.Equal(t, len(realElements), len(randomStatement.LexerElements))
		for i, element := range realElements {
			require.Equal(t, randomStatement.LexerElements[i].Color, element.Color)
		}
	})
}
