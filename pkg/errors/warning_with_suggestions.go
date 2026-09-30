package errors

type warningWithSuggestions struct {
	warnMsg        string
	suggestionsMsg string
}

func NewWarningWithSuggestions(warnMsg, suggestionsMsg string) *warningWithSuggestions {
	return &warningWithSuggestions{
		warnMsg:        warnMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func (w *warningWithSuggestions) DisplayWarningWithSuggestions() string {
	if w.warnMsg != "" && w.suggestionsMsg != "" {
		lines := "[WARN] " + w.warnMsg + "\n"
		lines += ComposeSuggestionsMessage(w.suggestionsMsg) + "\n"
		return lines
	}
	return ""
}
