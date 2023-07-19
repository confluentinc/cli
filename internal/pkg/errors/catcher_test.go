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
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"There is an error"}]}`)), StatusCode: http.StatusMethodNotAllowed}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, StatusCode(err))
}

func TestCatchErrorCodeWhenErrorMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)), StatusCode: http.StatusUnauthorized}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, http.StatusUnauthorized, StatusCode(err))
}

func TestCatchErrorCodeWhenNestedMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"error":{"message":"gateway error"}}`)), StatusCode: http.StatusMethodNotAllowed}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, StatusCode(err))
}

func TestCatchErrorOnlyStatusCode(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader("")), StatusCode: http.StatusMethodNotAllowed}

	err := CatchCCloudV2Error(errors.New("Some Error"), res)
	require.Error(t, err)
	require.Equal(t, http.StatusMethodNotAllowed, StatusCode(err))
}

func TestCatchServiceAccountExceedError(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 1000 service accounts"}]}`))}

	err := CatchServiceNameInUseError(New("402 Payment Required"), res, "")
	require.Error(t, err)
	require.Equal(t, "Your environment is currently limited to 1000 service accounts: 402 Payment Required", err.Error())
}
