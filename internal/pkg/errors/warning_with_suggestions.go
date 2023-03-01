package errors

import (
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const (
	warningsMessageHeader = "Warning: "
	reasonMessageHeader   = "Reason: "
)

type WarningWithSuggestions struct {
	warnMsg        string
	reasonMsg      string
	suggestionsMsg string
}

func NewWarningWithSuggestions(warnMsg string, reasonMsg string, suggestionsMsg string) *WarningWithSuggestions {
	return &WarningWithSuggestions{
		warnMsg:        warnMsg,
		reasonMsg:      reasonMsg,
		suggestionsMsg: suggestionsMsg,
	}
}

func (w *WarningWithSuggestions) DisplayWarningWithSuggestions() {
	if w.warnMsg != "" && w.reasonMsg != "" && w.suggestionsMsg != "" {
		utils.ErrPrintln(warningsMessageHeader + w.warnMsg)
		utils.ErrPrintln()
		utils.ErrPrintln(reasonMessageHeader + w.reasonMsg)
		utils.ErrPrintln(ComposeSuggestionsMessage(w.suggestionsMsg))
	}
}
