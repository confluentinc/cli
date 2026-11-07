//go:build eval

package scenarios

import "github.com/confluentinc/cli/v4/test/eval"

var selectionCrossField = eval.Scenario{
	Name:        "selection-cross-field",
	Description: "two sessions set DIFFERENT `current_*` fields on one shared context (`kafka cluster use` vs `flink compute-pool use`); a lockless whole-file save drops one.",
	Sessions: func(cloudURL string) []eval.SessionScript {
		return []eval.SessionScript{
			{Label: "uses kafka cluster", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"kafka cluster use lkc-12345"}},
			{Label: "uses flink compute-pool", Setup: []string{loginStep(cloudURL), "environment use " + envA}, Contend: []string{"flink compute-pool use lfcp-123456"}},
		}
	},
	Grade: func(results []eval.SessionResult) []eval.SessionOutcome {
		return []eval.SessionOutcome{
			eval.GradeSession(results[0], "lkc-12345", func(r eval.SessionResult) (string, error) { return eval.ReadActiveKafkaCluster(r.HomeDir, envA) }),
			eval.GradeSession(results[1], "lfcp-123456", func(r eval.SessionResult) (string, error) { return eval.ReadCurrentFlinkComputePool(r.HomeDir, envA) }),
		}
	},
}
