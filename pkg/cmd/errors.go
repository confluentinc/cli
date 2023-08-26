package cmd

import (
	"os"

	"github.com/confluentinc/cli/v3/pkg/color"
	"github.com/confluentinc/cli/v3/pkg/errors"
)

func CheckErr(enableColor bool, msg interface{}) {
	if msg == nil {
		return
	}

	color.ErrPrintf(enableColor, "Error: %v\n", msg)

	if err, ok := msg.(errors.ErrorWithSuggestions); ok {
		if suggestion := err.GetSuggestionsMsg(); suggestion != "" {
			color.ErrPrint(enableColor, errors.ComposeSuggestionsMessage(suggestion))
		}
	}

	os.Exit(1)
}
