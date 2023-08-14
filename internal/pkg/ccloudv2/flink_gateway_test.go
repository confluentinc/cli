package ccloudv2

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/errors/flink"
)

func TestFlinkErrorCodeWhenErrors(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"There is an error"}]}`)), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(errors.New("Some Error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.FlinkError)
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
	err := flink.CatchError(errors.New("Some Error"), nil)
	require.Error(t, err)

	flinkError, ok := err.(flink.FlinkError)
	require.True(t, ok)
	require.Equal(t, 0, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorCodeWhenErrorMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"message":"unauthorized"}`)), StatusCode: http.StatusUnauthorized}

	err := flink.CatchError(errors.New("Some Error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.FlinkError)
	require.True(t, ok)
	require.Equal(t, http.StatusUnauthorized, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorCodeWhenNestedMessage(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"error":{"message":"gateway error"}}`)), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(errors.New("Some Error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.FlinkError)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}

func TestFlinkErrorOnlyStatusCode(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader("")), StatusCode: http.StatusMethodNotAllowed}

	err := flink.CatchError(errors.New("Some Error"), res)
	require.Error(t, err)

	flinkError, ok := err.(flink.FlinkError)
	require.True(t, ok)
	require.Equal(t, http.StatusMethodNotAllowed, flinkError.StatusCode())
	require.Equal(t, err.Error(), flinkError.Error())
}
