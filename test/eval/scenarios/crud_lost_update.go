//go:build eval

package scenarios

import "github.com/confluentinc/cli/v4/test/eval"

var crudLostUpdate = eval.Scenario{
	Name:        "crud-lost-update",
	Description: "two sessions each `api-key create --resource global` on one shared context; a lockless whole-file save drops one of the two created keys.",
	Sessions: func(cloudURL string) []eval.SessionScript {
		return []eval.SessionScript{
			{Label: "creates key A", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
			{Label: "creates key B", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
		}
	},
	Grade: func(results []eval.SessionResult) []eval.SessionOutcome {
		outs := make([]eval.SessionOutcome, len(results))
		for i, r := range results {
			key := createdGlobalKey(r)
			outs[i] = eval.GradeSession(r, key, observeGlobalKeyPresent(key))
		}
		return outs
	},
}
