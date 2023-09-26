package autocomplete

import (
	"fmt"

	"github.com/confluentinc/go-prompt"

	"github.com/confluentinc/cli/v3/pkg/flink/config"
)

func SetCompleter(in prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: fmt.Sprintf("SET '%s' = '10000';", config.ConfigKeyResultsTimeout), Description: "Total amount of time in milliseconds to wait before timing out the request waiting for results to be ready."},
		{Text: fmt.Sprintf("SET '%s' = 'Europe/Berlin';", config.ConfigKeyLocalTimeZone), Description: "Used to set the timezone for the current session either with a TZID ('Europe/Berlin'), a fixed offset ('GMT+02:00') or just 'UTC'"},
	}

	return SuggestFromPrefix(s, in.TextBeforeCursor())
}
