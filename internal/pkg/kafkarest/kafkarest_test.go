package kafkarest

import (
	"fmt"
	"net/http"
	neturl "net/url"
	"testing"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func TestNewError(t *testing.T) {
	req := require.New(t)
	url := "http://my-url"
	neturlMsg := "net-error"

	neturlError := neturl.Error{
		Op:  "my-op",
		URL: url,
		Err: fmt.Errorf(neturlMsg),
	}

	r := NewError(url, &neturlError, nil)
	req.NotNil(r)
	req.Contains(r.Error(), "establish")
	req.Contains(r.Error(), url)
	req.Contains(r.Error(), neturlMsg)

	neturlError.Err = fmt.Errorf(SelfSignedCertError)
	r = NewError(url, &neturlError, nil)
	req.NotNil(r)
	req.Contains(r.Error(), "establish")
	req.Contains(r.Error(), url)
	e, ok := r.(errors.ErrorWithSuggestions)
	req.True(ok)
	req.Contains(e.GetSuggestionsMsg(), "CONFLUENT_PLATFORM_CA_CERT_PATH")

	openAPIError := kafkarestv3.GenericOpenAPIError{}

	r = NewError(url, openAPIError, nil)
	req.NotNil(r)
	req.Contains(r.Error(), "unknown")

	httpResp := http.Response{
		Status:     "Code: 400",
		StatusCode: http.StatusBadRequest,
		Request: &http.Request{
			Method: http.MethodGet,
			URL: &neturl.URL{
				Host: "myhost",
				Path: "/my-path",
			},
		},
	}
	r = NewError(url, openAPIError, &httpResp)
	req.NotNil(r)
	req.Contains(r.Error(), "failed")
	req.Contains(r.Error(), http.MethodGet)
	req.Contains(r.Error(), "myhost")
	req.Contains(r.Error(), "my-path")
}
