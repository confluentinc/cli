//go:build eval

package scenarios

import "github.com/confluentinc/cli/v4/test/eval"

var mixedWorkload = eval.Scenario{
	Name:        "mixed-workload",
	Description: "three sessions doing unrelated real work on one shared context (select an env; create an api-key; use a cluster then log out); under shared state one session's work clobbers another's.",
	Sessions: func(cloudURL string) []eval.SessionScript {
		return []eval.SessionScript{
			{Label: "selects an environment", Setup: []string{loginStep(cloudURL)}, Contend: []string{"environment use " + envA}},
			{Label: "creates an api-key", Setup: []string{loginStep(cloudURL)}, Contend: []string{"api-key create --resource global"}},
			{Label: "uses a cluster then logs out", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"kafka cluster use lkc-12345", "logout"}},
		}
	},
	Grade: func(results []eval.SessionResult) []eval.SessionOutcome {
		key := createdGlobalKey(results[1])
		return []eval.SessionOutcome{
			eval.GradeSession(results[0], envA, observeLoggedInEnvironment),
			eval.GradeSession(results[1], key, observeGlobalKeyPresent(key)),
			eval.GradeSession(results[2], "cleared", observeCredsCleared()),
		}
	},
}
