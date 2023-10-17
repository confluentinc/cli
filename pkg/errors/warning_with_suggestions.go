package errors

type WarningWithSuggestions struct {
	warnMsg        string
	reasonMsg      string
	suggestionsMsg string
}

func NewWarningWithSuggestions(warnMsg, reasonMsg, suggestionsMsg string) *WarningWithSuggestions {
	return &WarningWithSuggestions{
		warnMsg:        warnMsg,
		reasonMsg:      reasonMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func (w *WarningWithSuggestions) DisplayWarningWithSuggestions() string {
	if w.warnMsg != "" && w.reasonMsg != "" && w.suggestionsMsg != "" {
		lines := "[WARN] " + w.warnMsg + "\n"
		lines += "\n"
		lines += "Reason: " + w.reasonMsg + "\n"
		lines += ComposeSuggestionsMessage(w.suggestionsMsg) + "\n"
		return lines
	}
	return ""
}
