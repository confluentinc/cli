package usage

import (
	"testing"
)

func TestCollectAgentDetect_DoesNotPanic(t *testing.T) {
	u := New("1.2.3")
	u.CollectAgentDetect()
}
