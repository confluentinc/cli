package autocomplete

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/confluentinc/go-prompt"
)

type SuggestionFilterTestSuite struct {
	suite.Suite
}

func TestSuggestionFilterTestSuite(t *testing.T) {
	suite.Run(t, new(SuggestionFilterTestSuite))
}

func (s *SuggestionFilterTestSuite) TestGetLastWord() {
	assert.Equal(s.T(), "second", getLastWord("first second"))
	assert.Equal(s.T(), "second", getLastWord(" first second "))
	assert.Equal(s.T(), "second", getLastWord("    first second    "))
	assert.Equal(s.T(), "second", getLastWord("    first\tsecond    "))
	assert.Equal(s.T(), "second", getLastWord("    first\nsecond    "))
	assert.Equal(s.T(), "first", getLastWord("first"))
	assert.Equal(s.T(), "first", getLastWord("first "))
	assert.Equal(s.T(), "first", getLastWord("    first    "))
	assert.Equal(s.T(), "first", getLastWord("    first   "))
	assert.Equal(s.T(), "", getLastWord(""))
	assert.Equal(s.T(), "", getLastWord("  "))
}

func (s *SuggestionFilterTestSuite) TestGetNextWord() {
	assert.Equal(s.T(), "first", getNextWord("first second"))
	assert.Equal(s.T(), "first", getNextWord(" first second"))
	assert.Equal(s.T(), "first", getNextWord("    first second"))
	assert.Equal(s.T(), "first", getNextWord("\tfirst second"))
	assert.Equal(s.T(), "first", getNextWord("\nfirst second"))
	assert.Equal(s.T(), "first", getNextWord("first"))
	assert.Equal(s.T(), "first", getNextWord("first "))
	assert.Equal(s.T(), "first", getNextWord("    first "))
	assert.Equal(s.T(), "first", getNextWord("    first"))
	assert.Equal(s.T(), "", getNextWord(""))
	assert.Equal(s.T(), "", getNextWord("  "))
}

func (s *SuggestionFilterTestSuite) TestIsLastCharSpace() {
	assert.False(s.T(), isLastCharSpace(""))
	assert.False(s.T(), isLastCharSpace("Test"))
	assert.True(s.T(), isLastCharSpace(" "))
	assert.True(s.T(), isLastCharSpace("Test "))
	assert.True(s.T(), isLastCharSpace("Test  "))
	assert.True(s.T(), isLastCharSpace("Test\n"))
	assert.True(s.T(), isLastCharSpace("Test\t"))
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldNotIncludeCompleteWord() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, "SELECT * FROM ")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "Users WHERE userId = '123';", suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldIncludeIncompleteWord() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, "SELECT * FRO")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)

	suggestions = SuggestFromPrefix(prompts, "SELECT * FROM")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldGiveAllSuggestionsForEmptyPrefix() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, "")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), prompts[0].Text, suggestions[0].Text)
	assert.Equal(s.T(), prompts[1].Text, suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldGiveNoResultForMissingPrefix() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, " ")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = SuggestFromPrefix(prompts, "CREATE")
	assert.Equal(s.T(), 0, len(suggestions))
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldIgnoreCase() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, "select * fr")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)

	suggestions = SuggestFromPrefix(prompts, "SELECT * FR")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)

	suggestions = SuggestFromPrefix(prompts, "sElEct * FR")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestFromPrefixShouldIgnoreLineBreaks() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM Orders WHERE amount = 2;"},
		{Text: "SELECT * FROM Users WHERE userId = '123';"},
	}
	suggestions := SuggestFromPrefix(prompts, "select *\nfr")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)

	suggestions = SuggestFromPrefix(prompts, "SELECT *\n")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "FROM Orders WHERE amount = 2;", suggestions[0].Text)
	assert.Equal(s.T(), "FROM Users WHERE userId = '123';", suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldNotIncludeCompleteWord() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "TABLE ")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "Orders", suggestions[0].Text)
	assert.Equal(s.T(), "Orders;", suggestions[1].Text)

	suggestions = SuggestNextWord(prompts, "CREATE ")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "TABLE", suggestions[0].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldIncludeIncompleteWord() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "TABL")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
	assert.Equal(s.T(), "TABLE Orders;", suggestions[1].Text)

	suggestions = SuggestNextWord(prompts, "CREATE")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "CREATE TABLE", suggestions[0].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldIgnoreCase() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "TABL")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
	assert.Equal(s.T(), "TABLE Orders;", suggestions[1].Text)

	suggestions = SuggestNextWord(prompts, "tABl")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
	assert.Equal(s.T(), "TABLE Orders;", suggestions[1].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldIgnoreLineBreaks() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "\nTABL")
	assert.Equal(s.T(), 2, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
	assert.Equal(s.T(), "TABLE Orders;", suggestions[1].Text)

	suggestions = SuggestNextWord(prompts, "CREATE\n")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "TABLE", suggestions[0].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldGiveNoResultForEmptyWord() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "")
	assert.Equal(s.T(), 0, len(suggestions))

	suggestions = SuggestNextWord(prompts, " ")
	assert.Equal(s.T(), 0, len(suggestions))
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldGiveNoResultForMissingWord() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders;"},
	}
	suggestions := SuggestNextWord(prompts, "missing_word")
	assert.Equal(s.T(), 0, len(suggestions))
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldGiveNoDuplicates() {
	prompts := []prompt.Suggest{
		{Text: "CREATE TABLE Orders (product STRING, amount INT);"},
		{Text: "DROP TABLE Orders"},
	}
	suggestions := SuggestNextWord(prompts, "TABLE")
	assert.Equal(s.T(), 1, len(suggestions))
	assert.Equal(s.T(), "TABLE Orders", suggestions[0].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldRemoveParentheses() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM (Orders WHERE id = 1);"},
		{Text: "SELECT * FROM (Orders WHERE id = 1)"},
		{Text: "SELECT * FROM ( Orders WHERE id = 1 )"},
	}
	suggestions := SuggestNextWord(prompts, "1")
	assert.Equal(s.T(), 3, len(suggestions))
	assert.Equal(s.T(), "1;", suggestions[0].Text)
	assert.Equal(s.T(), "1", suggestions[1].Text)
	assert.Equal(s.T(), "1 ", suggestions[2].Text)
}

func (s *SuggestionFilterTestSuite) TestSuggestNextWordFromLastWordShouldNotReturnEmptySuggestion() {
	prompts := []prompt.Suggest{
		{Text: "SELECT * FROM ( Orders WHERE id = 1 )"},
	}
	suggestions := SuggestNextWord(prompts, "1 ")
	assert.Equal(s.T(), 0, len(suggestions))
}
