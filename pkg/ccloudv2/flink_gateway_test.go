package ccloudv2

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	flinkgatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/flink-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/errors/flink"
)

// The `confluent flink query` token refresh writes AuthToken from a background
// drain goroutine while a concurrent stop reads it (the read happens inside
// flinkGatewayApiContext on every gateway call). Run under -race, this fails if
// the token isn't serialized on the client. The command-level mocks can't cover
// this — they never touch the real client's token, which is where the read lives.
func TestFlinkGatewayClientAuthTokenIsRaceFree(t *testing.T) {
	client := NewFlinkGatewayClient("http://unused.invalid", "test", false, "initial")

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(3)
		go func(n int) { defer wg.Done(); client.SetAuthToken(fmt.Sprintf("token-%d", n)) }(i)
		go func() { defer wg.Done(); _ = client.GetAuthToken() }()
		// flinkGatewayApiContext is the real read path every gateway call takes.
		go func() {
			defer wg.Done()
			require.NotNil(t, client.flinkGatewayApiContext().Value(flinkgatewayv1.ContextAccessToken))
		}()
	}
	wg.Wait()
}

func TestFlinkErrorCodeWhenErrors(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"There is an error"}]}`)), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(fmt.Errorf("some error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.Error)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorNil(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"There is an error"}]}`)), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(nil, res)
	require.Nil(t, err)
}

func TestFlinkErrorNilHttpRes(t *testing.T) {
	err := flink.CatchError(fmt.Errorf("some error"), nil)
	require.Error(t, err)

	flinkError, ok := err.(flink.Error)
	require.True(t, ok)
	require.Equal(t, 0, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorCodeWhenErrorMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)), StatusCode: http.StatusUnauthorized}

	err := flink.CatchError(fmt.Errorf("some error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.Error)
	require.True(t, ok)
	require.Equal(t, http.StatusUnauthorized, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorCodeWhenNestedMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"error":{"message":"gateway error"}}`)), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(fmt.Errorf("some error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.Error)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorOnlyStatusCode(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader("")), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(fmt.Errorf("some error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.Error)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}
