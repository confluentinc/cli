package reverseisearch

import (
	"math"
	"strings"

	"github.com/confluentinc/go-prompt"
)

const BckISearch = "bck-i-search: "

type ReverseISearch interface {
	ReverseISearch(history []string) string
}

type reverseISearch struct{}

type SearchState struct {
	CurrentIndex int
	CurrentMatch string
}

type LivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

func NewReverseISearch() ReverseISearch {
	return reverseISearch{}
}

func reverseISearchLivePrefix(livePrefixState *LivePrefixState) func() (string, bool) {
	return func() (string, bool) {
		return livePrefixState.LivePrefix, livePrefixState.IsEnable
	}
}

func (r reverseISearch) ReverseISearch(history []string) string {
	writer := prompt.NewStdoutWriter()

	livePrefixState := &LivePrefixState{
		LivePrefix: BckISearch,
		IsEnable:   true,
	}

	reverseISearchEnabled := true
	searchState := &SearchState{
		CurrentIndex: len(history) - 1,
		CurrentMatch: "",
	}

	exitFromSearch := func(buffer *prompt.Buffer) {
		buffer.DeleteBeforeCursor(9999)
		reverseISearchEnabled = false
		livePrefixState.LivePrefix = ""
	}
	in := prompt.New(
		func(s string) {},
		searchCompleter(history, writer, searchState, livePrefixState),
		prompt.OptionSetExitCheckerOnInput(func(input string, lineBreak bool) bool {
			return !reverseISearchEnabled
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlC,
			Fn:  exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlM,
			Fn:  exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlQ,
			Fn:  exitFromSearch,
		}),
		prompt.OptionAddKeyBind(prompt.KeyBind{
			Key: prompt.ControlR,
			Fn:  nextResult(writer, history, searchState, livePrefixState),
		}),
		prompt.OptionWriter(writer),
		prompt.OptionTitle("bck-i-search"),
		prompt.OptionLivePrefix(reverseISearchLivePrefix(livePrefixState)),
		prompt.OptionHistory(history),
		prompt.OptionPrefixTextColor(prompt.DefaultColor),
		prompt.OptionSetStatementTerminator(func(lastKeyStroke prompt.Key, buffer *prompt.Buffer) bool {
			if lastKeyStroke == prompt.ControlM {
				livePrefixState.LivePrefix = ""
				return true
			}
			return false
		}),
	)
	in.Run()
	return searchState.CurrentMatch
}

// The searchCompleter is writing on console the suggestions from the history, appending the
// `bck-i-search: ` string. It always returns an empty []prompt.Suggest, because we are not using the built-in suggest.
func searchCompleter(history []string, writer prompt.ConsoleWriter, searchState *SearchState, livePrefix *LivePrefixState) prompt.Completer {
	return func(document prompt.Document) []prompt.Suggest {
		// User selected the command or key binding for next match
		if document.LastKeyStroke() == prompt.Escape || document.LastKeyStroke() == prompt.ControlR {
			return []prompt.Suggest{}
		}
		// user inserted a char, search start from top again
		searchState.CurrentIndex = len(history) - 1

		updateSuggestion(history, document.Text, writer, searchState, livePrefix)

		return []prompt.Suggest{}
	}
}

// nextResult will update console text with the next match from the history.
func nextResult(writer prompt.ConsoleWriter, history []string, searchState *SearchState, livePrefix *LivePrefixState) func(buffer *prompt.Buffer) {
	return func(buffer *prompt.Buffer) {
		searchState.CurrentIndex--
		updateSuggestion(history, buffer.Text(), writer, searchState, livePrefix)
	}
}

func updateSuggestion(history []string, substr string, writer prompt.ConsoleWriter, searchState *SearchState, livePrefix *LivePrefixState) {
	clearCurrentSuggestion(writer, searchState)
	result := search(substr, history, searchState.CurrentIndex)
	writeSuggestion(writer, result.match, result.matchIndexStart, len(substr))
	updateSearchState(searchState, result.match, result.index)
	updateLivePrefix(result.match, substr, livePrefix)
}

// In case there are no matches and the search string is not empty we've "failed". This mimics the behaviour of
// reverse search on unix shell.
func updateSearchState(searchState *SearchState, currentMatch string, currentIndex int) {
	searchState.CurrentMatch = currentMatch
	searchState.CurrentIndex = currentIndex
}

// In case there are no matches and the search string is not empty we've "failed". This mimics the behaviour of
// reverse search on unix shell.
func updateLivePrefix(match string, substr string, livePrefix *LivePrefixState) {
	if match == "" && substr != "" {
		livePrefix.LivePrefix = "failing " + BckISearch
	} else {
		livePrefix.LivePrefix = BckISearch
	}
}

// clearCurrentSuggestion will clear current text from the terminal, using the console writer and previous state.
// The most general case is when we previous had a multi-line statement match, like:
// > select * from
//
//		table 1
//	 join table 2
//	 bck-i-search: {CURSOR_IS_HERE}
//
// To clear this screen we need to do 3 steps:
// 1. EraseLine + up -> To clear `bck-i-search: ` line
// 2. (EraseLine + up) * #suggestions-1 -> To clear previous suggestions minus first line
// 3. EraseLine + move cursor at beginning of line + 1 -> First line with '>'
func clearCurrentSuggestion(writer prompt.ConsoleWriter, searchState *SearchState) {
	// 1. EraseLine + up -> To clear `bck-i-search: ` line
	writer.EraseLine()
	writer.CursorUp(1)

	// 2. (EraseLine + up) * #suggestions-1 -> To clear previous suggestions minus first line
	if lines := newLineCount(searchState.CurrentMatch); lines > 0 {
		for i := lines; i > 0; i-- {
			writer.EraseLine()
			writer.CursorUp(1)
		}
	}

	// 3. EraseLine + move cursor at beginning of line + 1 -> First line with '>'
	writer.CursorBackward(9999)
	writer.CursorForward(1)
	writer.EraseEndOfLine()
}

// writeSuggestion writes the match string, adding a line break. If there was a match it will be highlighted.
func writeSuggestion(writer prompt.ConsoleWriter, match string, matchIndex int, matchLength int) {
	suggestion := match + "\n"
	writer.WriteStr(" ")
	if matchIndex != -1 && matchLength != 0 {
		writer.SetColor(prompt.DefaultColor, prompt.DefaultColor, false)
		writer.WriteStr(suggestion[:matchIndex])
		writer.SetColor(prompt.DefaultColor, prompt.LightGray, true)
		writer.WriteStr(suggestion[matchIndex : matchIndex+matchLength])
		writer.SetColor(prompt.DefaultColor, prompt.DefaultColor, false)
		writer.WriteStr(suggestion[matchIndex+matchLength:])
	} else {
		writer.WriteStr(suggestion)
	}
}

// searchResult represent a match in the history:
// index is the index of the match in the history array, -1 otherwise
// match is the matched string: history[index] = match
// matchIndexStart is the index where the match start, -1 otherwise. e.g. "search(tch, {matching}" return 2
type searchResult struct {
	index           int
	match           string
	matchIndexStart int
}

// search for substr in the s slice backwards starting from the startIndex if specified.
func search(substr string, s []string, startIndex int) searchResult {
	// We want our backward search to be case insensitive, since flink sql is case insensitive for keywords.
	substr = strings.ToUpper(substr)

	// strings.contains(.., "") always return true
	if substr == "" {
		return searchResult{-1, "", -1}
	}
	// if start > size, just use size
	upperBound := int(math.Min(float64(startIndex), float64(len(s)-1)))
	for i := upperBound; i >= 0; i-- {
		substrI := strings.ToUpper(s[i])
		if strings.Contains(substrI, substr) {
			return searchResult{i, s[i], strings.Index(substrI, substr)}
		}
	}
	return searchResult{-1, "", -1}
}

func newLineCount(s string) int {
	return strings.Count(s, "\n")
}
