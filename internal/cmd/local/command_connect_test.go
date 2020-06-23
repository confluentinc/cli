package local

import (
	"bytes"
	"io/ioutil"
	"net/http"
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
		Body: ioutil.NopCloser(bytes.NewReader([]byte(exampleJSON))),
	}

	out, err := formatJSONResponse(res)
	req.NoError(err)
	req.Equal(exampleFormattedJSON, out)
}

func TestFormatEmptyJSONResponse(t *testing.T) {
	req := require.New(t)

	res := &http.Response{
		Body: ioutil.NopCloser(bytes.NewReader([]byte{})),
	}

	out, err := formatJSONResponse(res)
	req.NoError(err)
	req.Equal("", out)
}
