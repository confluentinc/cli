package errors

type WarningWithSuggestions struct {
	warnMsg        string
	suggestionsMsg string
}

func NewWarningWithSuggestions(warnMsg, suggestionsMsg string) *WarningWithSuggestions {
	return &WarningWithSuggestions{
		warnMsg:        warnMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func (w *WarningWithSuggestions) DisplayWarningWithSuggestions() string {
	if w.warnMsg != "" && w.suggestionsMsg != "" {
		lines := "[WARN] " + w.warnMsg + "\n"
		lines += "\n"
		lines += ComposeSuggestionsMessage(w.suggestionsMsg) + "\n"
		return lines
	}
	return ""
}
