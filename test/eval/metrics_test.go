//go:build eval

package eval

import (
	"math"
	"testing"
)

func TestAggregateComputesCollisionRateAndPassCaretK(t *testing.T) {
	outcomes := []TrialOutcome{
		{Sessions: 2, Collisions: 1, ConfigCorruptions: 0, AllPassed: false},
		{Sessions: 2, Collisions: 0, ConfigCorruptions: 0, AllPassed: true},
	}

	m := Aggregate(outcomes)

	if want := 1.0 / 4.0; math.Abs(m.CollisionRate-want) > 1e-9 {
		t.Fatalf("collision rate = %v, want %v", m.CollisionRate, want)
	}
	if want := 1.0 / 2.0; math.Abs(m.PassCaretK-want) > 1e-9 {
		t.Fatalf("pass^k = %v, want %v", m.PassCaretK, want)
	}
	if m.Trials != 2 {
		t.Fatalf("trials = %d, want 2", m.Trials)
	}
}
