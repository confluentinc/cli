//go:build eval

package scenarios

import (
	"fmt"

	"github.com/confluentinc/cli/v4/test/eval"
)

var crosstalkEnvs = []string{envA, envB}

var environmentCrosstalk = eval.Scenario{
	Name:        "environment-crosstalk",
	Description: "concurrent `confluent login` + `confluent environment use` sessions; shared state lands the wrong active environment, isolated state does not.",
	Sessions: func(cloudURL string) []eval.SessionScript {
		scripts := make([]eval.SessionScript, len(crosstalkEnvs))
		for i, env := range crosstalkEnvs {
			scripts[i] = eval.SessionScript{
				Label:   fmt.Sprintf("uses %s", env),
				Setup:   []string{loginStep(cloudURL)},
				Contend: []string{"environment use " + env},
			}
		}
		return scripts
	},
	Grade: func(results []eval.SessionResult) []eval.SessionOutcome {
		outs := make([]eval.SessionOutcome, len(results))
		for i, r := range results {
			outs[i] = eval.GradeSession(r, crosstalkEnvs[i], observeActiveEnvironment)
		}
		return outs
	},
}
