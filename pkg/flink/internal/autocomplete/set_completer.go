package autocomplete

import (
	"fmt"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v4/pkg/flink/config"
)

func SetCompleterCommon(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: fmt.Sprintf("SET '%s' = '10000';", config.KeyResultsTimeout), Description: "Total amount of time in milliseconds to wait before timing out the request waiting for results to be ready."},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}

// Additional examples for cloud
func SetCompleterCloud(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: fmt.Sprintf("SET '%s' = 'Europe/Berlin';", config.KeyLocalTimeZone), Description: "Used to set the timezone for the current session either with a TZID ('Europe/Berlin'), a fixed offset ('GMT+02:00') or just 'UTC'"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
