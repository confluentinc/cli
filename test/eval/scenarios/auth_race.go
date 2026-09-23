//go:build eval

package scenarios

import "github.com/confluentinc/cli/v4/test/eval"

var authRace = eval.Scenario{
	Name:        "auth-race",
	Description: "one session keeps working (`environment use`) while another `logout`s on the shared context; the logout can clobber the first session's auth, or be resurrected by it.",
	Sessions: func(cloudURL string) []eval.SessionScript {
		return []eval.SessionScript{
			{Label: "stays logged in, selects env", Setup: []string{loginStep(cloudURL)}, Contend: []string{"environment use " + envA}},
			{Label: "logs out", Setup: []string{loginStep(cloudURL)}, Contend: []string{"logout"}},
		}
	},
	Grade: func(results []eval.SessionResult) []eval.SessionOutcome {
		return []eval.SessionOutcome{
			eval.GradeSession(results[0], envA, observeActiveEnvironment),
			eval.GradeSession(results[1], "cleared", observeCredsCleared()),
		}
	},
}
