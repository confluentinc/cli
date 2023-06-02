package local

import (
	"runtime"
	"testing"

	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
)

func TestStartContainer(t *testing.T) {
	if runtime.GOOS == "darwin" {
		s.T().Skip()
	}

	req := require.New(t)
	pool, err := dockertest.NewPool("")
	req.NoError(err)

	resource, err := pool.Run("confluentinc/confluent-local", "latest", []string{})
	req.NoError(err)

	t.Cleanup(func() {
		require.NoError(t, pool.Purge(resource), "failed to remove container")
	})
}
