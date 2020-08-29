package local

import (
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	exampleJSON          = "{ \"key\": \"val\" }"
	exampleFormattedJSON = "{\n  \"key\": \"val\"\n}"
)

func TestIsJSON(t *testing.T) {
	req := require.New(t)

	req.True(isJSON([]byte(exampleJSON)))
	req.False(isJSON([]byte("Hello, World!")))
}

func TestFormatJSONResponse(t *testing.T) {
	req := require.New(t)

	res := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader(exampleJSON)),
	}

	out, err := formatJSONResponse(res)
	req.NoError(err)
	req.Equal(exampleFormattedJSON, out)
}

func TestFormatEmptyJSONResponse(t *testing.T) {
	req := require.New(t)

	res := &http.Response{
		Body: ioutil.NopCloser(strings.NewReader("")),
	}

	out, err := formatJSONResponse(res)
	req.NoError(err)
	req.Equal("", out)
}
