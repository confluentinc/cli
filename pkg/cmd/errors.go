package cmd

import (
	"os"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func CheckErr(enableColor bool, msg interface{}) {
	if msg == nil {
		return
	}

	output.ErrPrintf(enableColor, "Error: %v\n", msg)

	if err, ok := msg.(errors.ErrorWithSuggestions); ok {
		if suggestion := err.GetSuggestionsMsg(); suggestion != "" {
			output.ErrPrint(enableColor, errors.ComposeSuggestionsMessage(suggestion))
		}
	}

	os.Exit(1)
}
