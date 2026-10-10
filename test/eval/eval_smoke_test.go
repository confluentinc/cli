//go:build eval

package eval

import "testing"

func TestEvalHarnessTagCompiles(t *testing.T) {
	if got := 1 + 1; got != 2 {
		t.Fatalf("sanity check failed: got %d", got)
	}
}
