package errors

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCatchServiceAccountExceedError(t *testing.T) {
	res := &http.Response{Body: io.NopCloser(strings.NewReader(`{"errors":[{"detail":"Your environment is currently limited to 1000 service accounts"}]}`))}

	err := CatchServiceNameInUseError(fmt.Errorf("402 Payment Required"), res, "")
	require.Error(t, err)
	require.Equal(t, "Your environment is currently limited to 1000 service accounts", err.Error())
}
