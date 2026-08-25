package usage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectAgentDetect(t *testing.T) {
	u := New("1.2.3")
	u.CollectAgentDetect()

	// Always-set regardless of environment: Interactive and AgentTables are
	// never empty (see agentdetect.Attributes), so their presence proves the
	// mapping ran rather than being skipped or silently panicking.
	require.NotNil(t, u.Interactive)
	require.NotNil(t, u.AgentTables)
}
