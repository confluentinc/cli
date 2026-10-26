package ccloudv2

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/iotest"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"

	rtcesqlv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const lightningInlineResult = `{
	"api_version": "query/v1",
	"kind": "Statement",
	"result": {
		"result_format": "JSON",
		"schema": {"columns": [{"name": "id", "type": {"type": "INTEGER", "nullable": true}}]},
		"data": [["1"], ["2"]]
	}
}`

// lightningTestServer answers every query with one status and body, and records each request.
type lightningTestServer struct {
	*httptest.Server
	requests atomic.Int32

	mu          sync.Mutex
	lastRequest *http.Request
	lastBody    rtcesqlv1.QueryV1QueryRequest
}

func (s *lightningTestServer) last() (*http.Request, rtcesqlv1.QueryV1QueryRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastRequest, s.lastBody
}

func newLightningTestServer(t *testing.T, status int, contentType, body string) *lightningTestServer {
	server := &lightningTestServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.requests.Add(1)
		var received rtcesqlv1.QueryV1QueryRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&received))
		server.mu.Lock()
		server.lastRequest, server.lastBody = r.Clone(context.Background()), received
		server.mu.Unlock()

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("X-Request-Id", "a1b2c3d4")
		w.WriteHeader(status)
		_, err := w.Write([]byte(body))
		require.NoError(t, err)
	}))
	t.Cleanup(server.Close)
	return server
}

func newTestLightningClient(server *lightningTestServer) *LightningClient {
	client := NewLightningClient(server.URL, "test-agent", false, "dataplane-token")

	// Keep retries fast.
	retryableClient := client.GetConfig().HTTPClient.Transport.(*retryablehttp.RoundTripper).Client
	retryableClient.RetryWaitMin = time.Millisecond
	retryableClient.RetryWaitMax = time.Millisecond

	return client
}

func TestLightningExecuteQuery(t *testing.T) {
	server := newLightningTestServer(t, http.StatusOK, "application/json", lightningInlineResult)

	result, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT id FROM t", 0)
	require.NoError(t, err)

	request, body := server.last()
	require.Equal(t, http.MethodPost, request.Method)
	require.Equal(t, "/query/v1", request.URL.Path)
	require.Equal(t, "true", request.Header.Get(lightningSyncOnlyHeader))
	require.Equal(t, "Bearer dataplane-token", request.Header.Get("Authorization"))
	require.Equal(t, "test-agent", request.Header.Get("User-Agent"))

	require.Equal(t, "env-123456", body.GetCatalogName())
	require.Equal(t, "lkc-123456", body.GetDatabaseName())
	require.Equal(t, "SELECT id FROM t", body.GetQuery())
	require.Nil(t, body.Options, "no row limit is sent without a cap")

	require.Equal(t, "JSON", result.GetResultFormat())
	require.Len(t, result.Schema.GetColumns(), 1)
	require.NotNil(t, result.Data.ArrayOfQueryV1ResultRow)
	require.Len(t, *result.Data.ArrayOfQueryV1ResultRow, 2)
}

func TestLightningExecuteQueryMaxResultRows(t *testing.T) {
	server := newLightningTestServer(t, http.StatusOK, "application/json", lightningInlineResult)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT id FROM t", 101)
	require.NoError(t, err)
	_, body := server.last()
	require.NotNil(t, body.Options)
	require.Equal(t, int32(101), body.Options.GetMaxResultRows())
}

func TestLightningExecuteQueryTimeout(t *testing.T) {
	// A server that never answers, until the test ends.
	done := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		<-done
	}))
	t.Cleanup(func() {
		close(done)
		server.Close()
	})

	client := NewLightningClient(server.URL, "test-agent", false, "dataplane-token")
	client.queryTimeout = 50 * time.Millisecond

	_, err := client.ExecuteQuery("env-123456", "lkc-123456", "SELECT id FROM t", 0)
	require.EqualError(t, err, "the query timed out waiting for a response")
}

func TestLightningExecuteQueryAsyncRequired(t *testing.T) {
	server := newLightningTestServer(t, http.StatusUnprocessableEntity, "application/json", `{"errors": [{"code": "FE_ASYNC_REQUIRED", "status": "422", "detail": "Query requires asynchronous execution, which the sync-only request disallows."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT * FROM t", 0)
	require.EqualError(t, err, "the query result is too large or too slow to return synchronously")

	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Contains(t, errorWithSuggestions.GetSuggestionsMsg(), `"LIMIT"`)
}

func TestLightningExecuteQueryUnprocessableEntity(t *testing.T) {
	// A 422 for any other reason keeps the server's message.
	server := newLightningTestServer(t, http.StatusUnprocessableEntity, "application/json", `{"errors": [{"code": "FE_SOMETHING_ELSE", "status": "422", "detail": "Something else went wrong."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT * FROM t", 0)
	require.EqualError(t, err, "Something else went wrong.")
}

func TestLightningExecuteQueryAccepted(t *testing.T) {
	// A 202 must not be reported as an empty result, whatever its body.
	for _, contentType := range []string{"application/json", "text/plain"} {
		server := newLightningTestServer(t, http.StatusAccepted, contentType, `{"api_version": "query/v1", "kind": "Statement", "result": {"statement_id": "statement_1b2c3d4e5f6a", "environment_id": "env-123456", "kafka_cluster_id": "lkc-123456", "topic_name": "t"}}`)

		_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT * FROM t", 0)
		require.EqualError(t, err, "the query result is too large or too slow to return synchronously", contentType)
	}
}

func TestLightningExecuteQueryBadRequest(t *testing.T) {
	server := newLightningTestServer(t, http.StatusBadRequest, "application/json", `{"errors": [{"code": "FE_SQL_VALIDATION_ERROR", "status": "400", "detail": "Table 'orders' not found in database 'lkc-123456'."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT * FROM orders", 0)
	require.EqualError(t, err, "Table 'orders' not found in database 'lkc-123456'.")
	_, ok := err.(errors.ErrorWithSuggestions)
	require.False(t, ok, "only server errors suggest contacting support")
}

func TestLightningExecuteQueryServerError(t *testing.T) {
	// A 500 isn't retried and includes the request ID.
	server := newLightningTestServer(t, http.StatusInternalServerError, "application/json", `{"errors": [{"code": "FE_INTERNAL_ERROR", "status": "500", "detail": "The statement could not be completed."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "The statement could not be completed.")
	require.Equal(t, int32(1), server.requests.Load())

	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, `When contacting Confluent Support, include request ID "a1b2c3d4".`, errorWithSuggestions.GetSuggestionsMsg())
}

func TestLightningExecuteQueryServerErrorWithoutRequestId(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"errors": [{"code": "FE_INTERNAL_ERROR", "status": "500", "detail": "The statement could not be completed."}]}`))
	}))
	t.Cleanup(server.Close)

	_, err := NewLightningClient(server.URL, "test-agent", false, "dataplane-token").ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "The statement could not be completed.")
	_, ok := err.(errors.ErrorWithSuggestions)
	require.False(t, ok, "without a request ID, there is nothing to tell support")
}

func TestLightningExecuteQueryGatewayTimeout(t *testing.T) {
	// A gateway timeout isn't retried.
	server := newLightningTestServer(t, http.StatusServiceUnavailable, "text/plain; charset=utf-8", `{"errors": [{"status": "503", "code": "service_timeout", "detail": "Process took too long to complete. Try your request again later."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT * FROM t", 0)
	require.EqualError(t, err, "Process took too long to complete. Try your request again later.")
	require.Equal(t, int32(1), server.requests.Load())
}

func TestLightningExecuteQueryRestApiNotEnabled(t *testing.T) {
	server := newLightningTestServer(t, http.StatusForbidden, "application/json", `{"errors": [{"status": "403", "code": "api_not_enabled", "title": "Forbidden", "detail": "The REST API is not enabled for this organization."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "The REST API is not enabled for this organization.")
}

func TestLightningExecuteQueryErrorPage(t *testing.T) {
	// A non-JSON error body, such as a proxy's HTML page, is reported by its status.
	server := newLightningTestServer(t, http.StatusForbidden, "text/html", "<html><body>403 Forbidden</body></html>")

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "403 Forbidden")
}

func TestLightningExecuteQueryServerErrorPage(t *testing.T) {
	// A non-JSON 500 is reported by its status, with the request ID.
	server := newLightningTestServer(t, http.StatusInternalServerError, "text/html", "<html><body>500 Internal Server Error</body></html>")

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "500 Internal Server Error")

	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, `When contacting Confluent Support, include request ID "a1b2c3d4".`, errorWithSuggestions.GetSuggestionsMsg())
}

func TestLightningExecuteQueryRetriesExhausted(t *testing.T) {
	// After the last retry, the server's message is reported.
	server := newLightningTestServer(t, http.StatusTooManyRequests, "application/json", `{"errors": [{"code": "FE_RATE_LIMITED", "status": "429", "detail": "Too many requests. Retry after a short delay."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "Too many requests. Retry after a short delay.")
	require.Greater(t, server.requests.Load(), int32(1))
}

func TestLightningExecuteQueryUnreadableResult(t *testing.T) {
	// A value that isn't a string, array, or null can't be read.
	server := newLightningTestServer(t, http.StatusOK, "application/json", `{"api_version": "query/v1", "kind": "Statement", "result": {"result_format": "JSON", "schema": {"columns": []}, "data": [[1]]}}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.ErrorContains(t, err, "failed to read the query result: ")
}

func TestCatchLightningErrorWithoutResponse(t *testing.T) {
	dnsErr := &url.Error{Op: "Post", URL: "https://sql.us-east1.gcp.confluent.cloud/query/v1", Err: &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Err: "no such host", Name: "sql.us-east1.gcp.confluent.cloud", IsNotFound: true},
	}}
	err := catchLightningError(dnsErr, nil)
	require.EqualError(t, err, `no Lightning Tables endpoint found at "sql.us-east1.gcp.confluent.cloud"`)
	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, "Lightning Tables may not be available for the Kafka cluster's cloud provider or region.", errorWithSuggestions.GetSuggestionsMsg())

	// Other DNS failures, such as a timeout, are reported as is.
	dnsTimeout := &url.Error{Op: "Post", URL: "https://sql.us-west-2.aws.confluent.cloud/query/v1", Err: &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Err: "i/o timeout", Name: "sql.us-west-2.aws.confluent.cloud", IsTimeout: true},
	}}
	require.Equal(t, dnsTimeout, catchLightningError(dnsTimeout, nil))

	// Any other failure to get a response is reported as is.
	connectionErr := &url.Error{Op: "Post", URL: "https://sql.us-west-2.aws.confluent.cloud/query/v1", Err: errors.New("connection reset by peer")}
	require.Equal(t, connectionErr, catchLightningError(connectionErr, nil))
}

func TestIsLightningAsyncRequired(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		httpResp *http.Response
	}{
		{"no response", rtcesqlv1.GenericOpenAPIError{}, nil},
		{"not an API error", errors.New("connection reset"), &http.Response{StatusCode: http.StatusUnprocessableEntity}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.False(t, isLightningAsyncRequired(test.err, test.httpResp))
		})
	}

	// GenericOpenAPIError's body is unexported, so a real one comes from a request.
	execute := func(status int, body string) (error, *http.Response) {
		server := newLightningTestServer(t, status, "application/json", body)
		client := newTestLightningClient(server)
		_, httpResp, err := client.StatementsQueryV1Api.ExecuteQueryV1Statement(client.lightningApiContext()).QueryV1QueryRequest(*rtcesqlv1.NewQueryV1QueryRequest("env-123456", "lkc-123456", "SELECT 1")).Execute()
		return err, httpResp
	}

	err, httpResp := execute(http.StatusUnprocessableEntity, `{"errors": [{"code": "FE_OTHER"}, {"code": "FE_ASYNC_REQUIRED"}]}`)
	require.True(t, isLightningAsyncRequired(err, httpResp))

	err, httpResp = execute(http.StatusUnprocessableEntity, "not json")
	require.False(t, isLightningAsyncRequired(err, httpResp))

	// The code only counts on a 422.
	err, httpResp = execute(http.StatusBadRequest, `{"errors": [{"code": "FE_ASYNC_REQUIRED"}]}`)
	require.False(t, isLightningAsyncRequired(err, httpResp))
}

func TestLightningHttpClientRetries(t *testing.T) {
	roundTripper, ok := newLightningHttpClient(false).Transport.(*retryablehttp.RoundTripper)
	require.True(t, ok)
	checkRetry := roundTripper.Client.CheckRetry

	const gatewayTimeout = `{"errors": [{"status": "503", "code": "service_timeout", "detail": "Process took too long to complete. Try your request again later."}]}`
	tests := []struct {
		status int
		body   string
		want   bool
	}{
		{http.StatusOK, "", false},
		{http.StatusBadRequest, "", false},
		{http.StatusUnprocessableEntity, "", false},
		{http.StatusInternalServerError, "", false},
		{http.StatusTooManyRequests, "", true},
		{http.StatusBadGateway, "", true},
		{http.StatusServiceUnavailable, "<html>503 Service Unavailable</html>", true},
		{http.StatusServiceUnavailable, `{"errors": [{"status": "503", "detail": "RTCE service mTLS client is not ready"}]}`, true},
		// The gateway's time limit.
		{http.StatusServiceUnavailable, gatewayTimeout, false},
		{http.StatusGatewayTimeout, "", false},
	}
	for _, test := range tests {
		resp := &http.Response{StatusCode: test.status, Body: io.NopCloser(strings.NewReader(test.body))}
		retry, err := checkRetry(context.Background(), resp, nil)
		require.NoError(t, err)
		require.Equal(t, test.want, retry, "status %d, body %s", test.status, test.body)

		// The body stays readable.
		body, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		require.Equal(t, test.body, string(body))
	}

	// A 503 with an unreadable body is retried and keeps a readable body.
	resp := &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(iotest.ErrReader(errors.New("connection reset")))}
	retry, err := checkRetry(context.Background(), resp, nil)
	require.NoError(t, err)
	require.True(t, retry)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	require.Empty(t, body)

	retry, err = checkRetry(context.Background(), nil, errors.New("connection refused"))
	require.False(t, retry)
	require.Error(t, err)
}

func TestGetLightningUrl(t *testing.T) {
	tests := []struct {
		platformServer string
		cloud          string
		region         string
		want           string
	}{
		{"https://confluent.cloud", "AWS", "us-west-2", "https://sql.us-west-2.aws.confluent.cloud"},
		{"https://confluent.cloud", "gcp", "us-central1", "https://sql.us-central1.gcp.confluent.cloud"},
		{"https://stag.cpdev.cloud", "AZURE", "westus2", "https://sql.westus2.azure.stag.cpdev.cloud"},
		{"https://devel.cpdev.cloud/", "AWS", "us-west-2", "https://sql.us-west-2.aws.devel.cpdev.cloud"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			cfg := &config.Config{
				CurrentContext: "cloud",
				Contexts:       map[string]*config.Context{"cloud": {Platform: &config.Platform{Server: test.platformServer}}},
			}
			url, err := GetLightningUrl(cfg, test.cloud, test.region)
			require.NoError(t, err)
			require.Equal(t, test.want, url)
		})
	}

	url, err := GetLightningUrl(&config.Config{IsTest: true}, "AWS", "us-west-2")
	require.NoError(t, err)
	require.Equal(t, testserver.TestLightningUrl.String(), url)

	cfg := &config.Config{
		CurrentContext: "cloud",
		Contexts:       map[string]*config.Context{"cloud": {Platform: &config.Platform{Server: "https://bad host"}}},
	}
	_, err = GetLightningUrl(cfg, "AWS", "us-west-2")
	require.Error(t, err)
}
