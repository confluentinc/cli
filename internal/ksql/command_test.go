package ksql

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInfoHttpConfigRetryStatus(t *testing.T) {
	retryStatus := newInfoHttpConfig().RetryStatus
	tests := []struct {
		status int
		want   bool
	}{
		{http.StatusOK, false},
		{http.StatusBadRequest, false},
		{http.StatusUnauthorized, false},
		{http.StatusRequestTimeout, true},
		{http.StatusTooManyRequests, true},
		{http.StatusInternalServerError, true},
		{http.StatusBadGateway, true},
		// The answer /info gives for misconfigured ACLs (error code 50321).
		{http.StatusServiceUnavailable, false},
		{http.StatusGatewayTimeout, true},
	}
	for _, test := range tests {
		require.Equal(t, test.want, retryStatus(&http.Response{StatusCode: test.status}), "status %d", test.status)
	}
}
