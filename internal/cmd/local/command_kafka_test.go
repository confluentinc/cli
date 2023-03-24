package local

import (
	"testing"

	"github.com/ory/dockertest"
	"github.com/stretchr/testify/require"
)

func TestStartContainer(t *testing.T) {
	req := require.New(t)
	pool, err := dockertest.NewPool("")
	req.NoError(err)

	resource, err := pool.Run("523370736235.dkr.ecr.us-west-2.amazonaws.com/confluentinc/kafka-local", "latest", []string{}) // add to ci
	req.NoError(err)

	t.Cleanup(func() {
		require.NoError(t, pool.Purge(resource), "failed to remove container")
	})
}
