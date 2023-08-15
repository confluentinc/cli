package errors

import (
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

func TestCatchServiceAccountExceedError(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 1000 service accounts"}]}`))}

	err := CatchServiceNameInUseError(New("402 Payment Required"), res, "")
	require.Error(t, err)
	require.Equal(t, "Your environment is currently limited to 1000 service accounts: 402 Payment Required", err.Error())
}
