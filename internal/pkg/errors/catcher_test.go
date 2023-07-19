package errors

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatchClustersExceedError(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 50 kafka clusters"}]}`))}

	err := CatchClusterConfigurationNotValidError(New("402 Payment Required"), res)
	require.Error(t, err)
	require.Equal(t, err.Error(), "Your environment is currently limited to 50 kafka clusters: 402 Payment Required")
}

func TestCatchErrorCodeWhenErrors(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"There is an error"}]}`)), StatusCode: 405}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, 405, StatusCode(err))

}

func TestCatchErrorCodeWhenErrorMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)), StatusCode: 401}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, 401, StatusCode(err))

}

func TestCatchErrorCodeWhenNestedMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"error":{"message":"gateway error"}}`)), StatusCode: 405}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, 405, StatusCode(err))

}

func TestCatchErrorOnlyStatusCode(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader("")), StatusCode: 405}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, 405, StatusCode(err))

}

func TestCatchServiceAccountExceedError(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 1000 service accounts"}]}`))}

	err := CatchServiceNameInUseError(New("402 Payment Required"), res, "")
	require.Error(t, err)
	require.Equal(t, "Your environment is currently limited to 1000 service accounts: 402 Payment Required", err.Error())
}
