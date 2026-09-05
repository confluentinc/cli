package usage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCollectAgentDetect(t *testing.T) {
	u := New("1.2.3")
	u.CollectAgentDetect()

	// Always-set: Interactive and AgentTables (see agentdetect.Attributes)
	// are never empty, so their presence proves the mapping ran.
	require.NotNil(t, u.Interactive)
	require.NotNil(t, u.AgentTables)
}
