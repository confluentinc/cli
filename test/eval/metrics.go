//go:build eval

package eval

type TrialOutcome struct {
	Sessions          int
	Collisions        int
	ConfigCorruptions int
	AllPassed         bool
}

// GradeTrial scores one concurrent trial from its session results.
func GradeTrial(results []SessionResult) TrialOutcome {
	out := TrialOutcome{Sessions: len(results)}
	for _, r := range results {
		if err := GradeConfigIntegrity(r.HomeDir); err != nil {
			out.ConfigCorruptions++
			continue
		}
		ok, err := GradeTargetFidelity(r.HomeDir, r.IntendedEnv)
		if err != nil || !ok {
			out.Collisions++
		}
	}
	out.AllPassed = out.Collisions == 0 && out.ConfigCorruptions == 0
	return out
}

type CellMetrics struct {
	Trials         int
	CollisionRate  float64
	CorruptionRate float64
	PassCaretK     float64
}

// Aggregate rolls trial outcomes into per-cell rates.
func Aggregate(outcomes []TrialOutcome) CellMetrics {
	m := CellMetrics{Trials: len(outcomes)}
	if len(outcomes) == 0 {
		return m
	}
	var totalSessions, totalCollisions, totalCorruptions, allPassed int
	for _, o := range outcomes {
		totalSessions += o.Sessions
		totalCollisions += o.Collisions
		totalCorruptions += o.ConfigCorruptions
		if o.AllPassed {
			allPassed++
		}
	}
	if totalSessions > 0 {
		m.CollisionRate = float64(totalCollisions) / float64(totalSessions)
		m.CorruptionRate = float64(totalCorruptions) / float64(totalSessions)
	}
	m.PassCaretK = float64(allPassed) / float64(len(outcomes))
	return m
}
