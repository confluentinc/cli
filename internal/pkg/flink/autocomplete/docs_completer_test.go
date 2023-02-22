package autocomplete

import (
	"github.com/c-bata/go-prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type DocsCompleterTestSuite struct {
	suite.Suite
}

func TestDocsCompleterTestSuite(t *testing.T) {
	suite.Run(t, new(DocsCompleterTestSuite))
}

func (s *DocsCompleterTestSuite) TestSuggestNextWordFromLastWord() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
		{Text: "ALTER TABLE Orders RENAME TO NewOrders;"},
		{Text: "INSERT INTO Orders VALUES ('pen', 2);"},
		{Text: "ANALYZE TABLE Orders COMPUTE STATISTICS;"},
		{Text: "DESCRIBE Orders;"},
		{Text: "EXPLAIN PLAN FOR ...;"},
		{Text: "USE db1;"},
		{Text: "SHOW TABLES;"},
		{Text: "SET 'table.local-time-zone;' = 'Europe/Berlin';"},
		{Text: "RESET;"},
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SHOW (TABLES;)"},
	}
	suggestions := suggestNextWordFromLastWord(prompts, "")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = suggestNextWordFromLastWord(prompts, " ")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = suggestNextWordFromLastWord(prompts, "specific_field")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = suggestNextWordFromLastWord(prompts, "specific_field ")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = suggestNextWordFromLastWord(prompts, "FROM")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "FROM Orders", suggestions[0].Text)

	suggestions = suggestNextWordFromLastWord(prompts, "FROM ")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "Orders", suggestions[0].Text)

	suggestions = suggestNextWordFromLastWord(prompts, "fRo")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "FROM Orders", suggestions[0].Text)

	suggestions = suggestNextWordFromLastWord(prompts, "f")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FOR ...;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Orders", suggestions[1].Text)

	suggestions = suggestNextWordFromLastWord(prompts, "TA")
	assert.Equal(s.T(), 3, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
	assert.Equal(s.T(), "TABLE Orders;", suggestions[1].Text)
	assert.Equal(s.T(), "TABLES;", suggestions[2].Text)
}

func (s *DocsCompleterTestSuite) TestGetNextWord() {
	assert.Equal(s.T(), "first", getNextWord("first second"))
	assert.Equal(s.T(), "first", getNextWord(" first second"))
	assert.Equal(s.T(), "first", getNextWord("    first second"))
	assert.Equal(s.T(), "first", getNextWord("first"))
	assert.Equal(s.T(), "first", getNextWord("first "))
	assert.Equal(s.T(), "first", getNextWord("    first "))
	assert.Equal(s.T(), "first", getNextWord("    first"))
	assert.Equal(s.T(), "", getNextWord(""))
	assert.Equal(s.T(), "", getNextWord("  "))
}

func (s *DocsCompleterTestSuite) TestIsLastCharSpace() {
	assert.False(s.T(), isLastCharSpace(""))
	assert.True(s.T(), isLastCharSpace(" "))
	assert.False(s.T(), isLastCharSpace("Test"))
	assert.True(s.T(), isLastCharSpace("Test "))
	assert.True(s.T(), isLastCharSpace("Test  "))
}
