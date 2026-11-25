package httpclient

import (
	"net/http"
)

var userAgent string

func SetUserAgent(agent string) {
	userAgent = agent
}

type userAgentTransport struct {
	next http.RoundTripper
}

// Set User-Agent if not already set; v2 SDKs for example will set this while preparing the request
func (t *userAgentTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if userAgent != "" && req.Header.Get("User-Agent") == "" {
		req = req.Clone(req.Context())
		req.Header.Set("User-Agent", userAgent)
	}

	return t.next.RoundTrip(req)
}
