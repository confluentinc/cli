package ccloudv2

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/require"

	rtcesqlv1alpha1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1alpha1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const lightningInlineResult = `{
	"api_version": "query/v1alpha1",
	"kind": "Statement",
	"result": {
		"result_format": "JSON",
		"schema": {"columns": [{"name": "id", "type": {"type": "INTEGER", "nullable": true}}]},
		"data": [["1"], ["2"]]
	}
}`

// lightningTestServer answers every query with one status and body, and records what it received.
type lightningTestServer struct {
	*httptest.Server
	requests atomic.Int32

	mu          sync.Mutex
	lastRequest *http.Request
	lastBody    rtcesqlv1alpha1.QueryV1alpha1QueryRequest
}

func (s *lightningTestServer) last() (*http.Request, rtcesqlv1alpha1.QueryV1alpha1QueryRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastRequest, s.lastBody
}

func newLightningTestServer(t *testing.T, status int, contentType, body string) *lightningTestServer {
	server := &lightningTestServer{}
	server.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.requests.Add(1)
		var received rtcesqlv1alpha1.QueryV1alpha1QueryRequest
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

	// Keep retries fast; the retry policy itself is what's under test.
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
	require.Equal(t, "/query/v1alpha1", request.URL.Path)
	require.Equal(t, "true", request.Header.Get(lightningSyncOnlyHeader))
	require.Equal(t, "Bearer dataplane-token", request.Header.Get("Authorization"))
	require.Equal(t, "test-agent", request.Header.Get("User-Agent"))

	require.Equal(t, "env-123456", body.GetCatalogName())
	require.Equal(t, "lkc-123456", body.GetDatabaseName())
	require.Equal(t, "SELECT id FROM t", body.GetQuery())
	require.Nil(t, body.Options, "no row limit is sent without a cap")

	require.Equal(t, "JSON", result.GetResultFormat())
	require.Len(t, result.Schema.GetColumns(), 1)
	require.NotNil(t, result.Data.ArrayOfQueryV1alpha1ResultRow)
	require.Len(t, *result.Data.ArrayOfQueryV1alpha1ResultRow, 2)
}

func TestLightningExecuteQueryMaxResultRows(t *testing.T) {
	server := newLightningTestServer(t, http.StatusOK, "application/json", lightningInlineResult)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT id FROM t", 101)
	require.NoError(t, err)
	_, body := server.last()
	require.NotNil(t, body.Options)
	require.Equal(t, int32(101), body.Options.GetMaxResultRows())
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
	// The SDK decodes a 202 as an inline result without rows; it must not be reported as an empty
	// result, whatever its body.
	for _, contentType := range []string{"application/json", "text/plain"} {
		server := newLightningTestServer(t, http.StatusAccepted, contentType, `{"api_version": "query/v1alpha1", "kind": "Statement", "result": {"statement_id": "statement_1b2c3d4e5f6a", "environment_id": "env-123456", "kafka_cluster_id": "lkc-123456", "topic_name": "t"}}`)

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
	// A 500 often fails the same way every time, so it is reported at once, with the request ID.
	server := newLightningTestServer(t, http.StatusInternalServerError, "application/json", `{"errors": [{"code": "FE_INTERNAL_ERROR", "status": "500", "detail": "The statement could not be completed."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "The statement could not be completed.")
	require.Equal(t, int32(1), server.requests.Load())

	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, `When contacting Confluent Support, include request ID "a1b2c3d4".`, errorWithSuggestions.GetSuggestionsMsg())
}

func TestLightningExecuteQueryRetriesExhausted(t *testing.T) {
	// After the last retry, the server's message is reported rather than a generic "giving up" error.
	server := newLightningTestServer(t, http.StatusTooManyRequests, "application/json", `{"errors": [{"code": "FE_RATE_LIMITED", "status": "429", "detail": "Too many requests. Retry after a short delay."}]}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.EqualError(t, err, "Too many requests. Retry after a short delay.")
	require.Greater(t, server.requests.Load(), int32(1))
}

func TestLightningExecuteQueryUnreadableResult(t *testing.T) {
	// Values are always strings, arrays, or null; anything else means the result can't be read.
	server := newLightningTestServer(t, http.StatusOK, "application/json", `{"api_version": "query/v1alpha1", "kind": "Statement", "result": {"result_format": "JSON", "schema": {"columns": []}, "data": [[1]]}}`)

	_, err := newTestLightningClient(server).ExecuteQuery("env-123456", "lkc-123456", "SELECT 1", 0)
	require.ErrorContains(t, err, "failed to read the query result: ")
}

func TestCatchLightningErrorWithoutResponse(t *testing.T) {
	dnsErr := &url.Error{Op: "Post", URL: "https://sql.us-east1.gcp.confluent.cloud/query/v1alpha1", Err: &net.OpError{
		Op:  "dial",
		Net: "tcp",
		Err: &net.DNSError{Err: "no such host", Name: "sql.us-east1.gcp.confluent.cloud", IsNotFound: true},
	}}
	err := catchLightningError(dnsErr, nil)
	require.EqualError(t, err, `no Lightning Tables endpoint found at "sql.us-east1.gcp.confluent.cloud"`)
	errorWithSuggestions, ok := err.(errors.ErrorWithSuggestions)
	require.True(t, ok)
	require.Equal(t, "Lightning Tables may not be available for the Kafka cluster's cloud provider or region.", errorWithSuggestions.GetSuggestionsMsg())

	// Any other failure to get a response is reported as is.
	connectionErr := &url.Error{Op: "Post", URL: "https://sql.us-west-2.aws.confluent.cloud/query/v1alpha1", Err: errors.New("connection reset by peer")}
	require.Equal(t, connectionErr, catchLightningError(connectionErr, nil))
}

func TestIsLightningAsyncRequired(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		httpResp *http.Response
	}{
		{"no response", rtcesqlv1alpha1.GenericOpenAPIError{}, nil},
		{"other status", rtcesqlv1alpha1.GenericOpenAPIError{}, &http.Response{StatusCode: http.StatusBadRequest}},
		{"not an API error", errors.New("connection reset"), &http.Response{StatusCode: http.StatusUnprocessableEntity}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.False(t, isLightningAsyncRequired(test.err, test.httpResp))
		})
	}

	// GenericOpenAPIError's body is unexported, so a real one comes from a request.
	execute := func(body string) (error, *http.Response) {
		server := newLightningTestServer(t, http.StatusUnprocessableEntity, "application/json", body)
		client := newTestLightningClient(server)
		_, httpResp, err := client.StatementsQueryV1alpha1Api.ExecuteQueryV1alpha1Statement(client.lightningApiContext()).QueryV1alpha1QueryRequest(*rtcesqlv1alpha1.NewQueryV1alpha1QueryRequest("env-123456", "lkc-123456", "SELECT 1")).Execute()
		return err, httpResp
	}

	err, httpResp := execute(`{"errors": [{"code": "FE_OTHER"}, {"code": "FE_ASYNC_REQUIRED"}]}`)
	require.True(t, isLightningAsyncRequired(err, httpResp))

	err, httpResp = execute("not json")
	require.False(t, isLightningAsyncRequired(err, httpResp))
}

func TestLightningHttpClientRetries(t *testing.T) {
	roundTripper, ok := newLightningHttpClient(false).Transport.(*retryablehttp.RoundTripper)
	require.True(t, ok)
	checkRetry := roundTripper.Client.CheckRetry

	for status, want := range map[int]bool{
		http.StatusOK:                  false,
		http.StatusBadRequest:          false,
		http.StatusUnprocessableEntity: false,
		http.StatusInternalServerError: false,
		http.StatusTooManyRequests:     true,
		http.StatusBadGateway:          true,
		http.StatusServiceUnavailable:  true,
		http.StatusGatewayTimeout:      true,
	} {
		retry, err := checkRetry(context.Background(), &http.Response{StatusCode: status}, nil)
		require.NoError(t, err)
		require.Equal(t, want, retry, "status %d", status)
	}

	retry, err := checkRetry(context.Background(), nil, errors.New("connection refused"))
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
