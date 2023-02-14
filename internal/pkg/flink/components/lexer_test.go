package components

import (
	"strings"
	"testing"

	prompt "github.com/c-bata/go-prompt"
)

func TestLexer(t *testing.T) {
	// given
	line := "SELECT FIELD FROM TABLE;"

	// when
	elements := lexer(line)

	// then
	for i, element := range elements {
		if i >= 0 && i < 6 || i > 12 && i < 17 {
			if element.Color != HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, HIGHLIGHT_COLOR)
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
	elements := lexer(line)

	// then
	for i, element := range elements {
		if i >= 0 && i < 6 || i > 12 && i < 17 {
			if element.Color != HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, HIGHLIGHT_COLOR)
			}
		} else if element.Color != prompt.White {
			t.Errorf("lexer() = %d, want %d", element.Color, prompt.White)
		}

	}
}

func TestWordLexer(t *testing.T) {
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
		elements := wordLexer(statement)

		// then
		for _, element := range elements {

			_, isKeyWord := SQLKeywords[strings.TrimSpace(element.Text)]

			if isKeyWord && element.Color != HIGHLIGHT_COLOR {
				t.Errorf("lexer() = %d, want %d", element.Color, HIGHLIGHT_COLOR)
			} else if !isKeyWord && element.Color != prompt.White {
				t.Errorf("lexer() = %d, want %d", element.Color, prompt.White)
			}

		}
	}
}
