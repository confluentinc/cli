package controller

import (
	"strings"

	"github.com/confluentinc/go-prompt"
)

const reverseISearch = "bck-i-search: "

type SearchState struct {
	index        int
	currentMatch string
}

type ReverseISearchLivePrefixState struct {
	LivePrefix string
	IsEnable   bool
}

// The reverseISearchCompleter is writing on console the suggestions from the history, appending the
// `bck-i-search: ` string. It always returns an empty []prompt.Suggest, because we are not using the built-in suggest.
func reverseISearchCompleter(history []string, writer prompt.ConsoleWriter, searchState *SearchState, livePrefix *ReverseISearchLivePrefixState) prompt.Completer {
	return func(document prompt.Document) []prompt.Suggest {

		// User selected the command
		if document.LastKeyStroke() == prompt.Escape {
			return []prompt.Suggest{}
		}

		// `bck-i-search: ' is part of the input document, need to strip it out
		substr := strings.Replace(document.Text, reverseISearch, "", -1)

		clearCurrentSuggestion(writer, searchState)
		match := writeSuggestion(writer, substr, history)

		updateLivePrefix(match, substr, livePrefix)

		searchState.currentMatch = match

		return []prompt.Suggest{}
	}
}

// In case there are no matches and the search string is not empty we've "failed". This mimics the behaviour of
// reverse search on unix shell.
func updateLivePrefix(match string, substr string, livePrefix *ReverseISearchLivePrefixState) {
	if match == "" && substr != "" {
		livePrefix.LivePrefix = "failing " + reverseISearch
	} else {
		livePrefix.LivePrefix = reverseISearch
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
	if lines := NewLineCount(searchState.currentMatch); lines > 0 {
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

// Write the suggestion to output, adding a line break and the string `"bck-i-search: "` in case of success.
// todo: add `failing` string before `bck-i-search: ` string.
func writeSuggestion(writer prompt.ConsoleWriter, substr string, history []string) string {

	searchResult, _ := search(substr, history)

	suggestion := searchResult + "\n"

	writer.WriteStr(suggestion)

	return searchResult
}

// Matches input string backward in the history, and returns the starting index of the match. If no
// match returns "" and -1.
func search(substr string, history []string) (string, int) {
	match := ""
	if substr == "" {
		return match, -1
	}
	for i := len(history) - 1; i >= 0; i-- {
		if strings.Contains(history[i], substr) {
			return history[i], strings.Index(history[i], substr)
		}
	}
	return match, -1
}

func NewLineCount(s string) int {
	return strings.Count(s, "\n")
}
