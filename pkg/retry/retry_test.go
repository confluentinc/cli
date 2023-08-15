package retry

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRetry(t *testing.T) {
	require.Error(t, Retry(time.Nanosecond, 2*time.Nanosecond, func() error {
		return errors.New("error")
	}))
	require.NoError(t, Retry(time.Nanosecond, 2*time.Nanosecond, func() error {
		return nil
	}))
}