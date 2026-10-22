package ccloudv2

import (
	"bytes"
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-retryablehttp"

	rtcesqlv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/rtce-sql/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	plog "github.com/confluentinc/cli/v4/pkg/log"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const (
	// lightningSyncOnlyHeader makes the query endpoint answer inline or fail. Where asynchronous queries
	// are enabled, a query whose result is too large or too slow to return inline is otherwise accepted as
	// an asynchronous job (HTTP 202), which this CLI does not support; with the header, that query fails
	// with HTTP 422 and lightningAsyncRequiredCode instead. Elsewhere, such a query runs inline until it
	// finishes or reaches the API gateway's time limit.
	lightningSyncOnlyHeader = "X-Lightning-Sync-Only"

	lightningAsyncRequiredCode = "FE_ASYNC_REQUIRED"

	// lightningGatewayTimeoutCode is the error code of the HTTP 503 the API gateway returns when a
	// request runs into its time limit.
	lightningGatewayTimeoutCode = "service_timeout"
)

type LightningClient struct {
	*rtcesqlv1.APIClient
	authToken string
}

func NewLightningClient(url, userAgent string, unsafeTrace bool, authToken string) *LightningClient {
	cfg := rtcesqlv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.DefaultHeader = map[string]string{lightningSyncOnlyHeader: "true"}
	cfg.HTTPClient = newLightningHttpClient(unsafeTrace)
	cfg.Servers = rtcesqlv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &LightningClient{
		APIClient: rtcesqlv1.NewAPIClient(cfg),
		authToken: authToken,
	}
}

// newLightningHttpClient retries only throttling and gateway errors. Unlike the other APIs, the query
// endpoint also returns HTTP 500 for queries that fail the same way on every attempt, so retrying a 500
// would only delay the error. A gateway timeout (HTTP 504, or a 503 with lightningGatewayTimeoutCode)
// isn't retried either: the query would most likely run into the time limit again. When retries run
// out, the last response is returned as is, so the user sees the server's message rather than a
// generic one.
func newLightningHttpClient(unsafeTrace bool) *http.Client {
	client := retryablehttp.NewClient()
	client.Logger = plog.NewLeveledLogger(unsafeTrace)
	client.CheckRetry = func(_ context.Context, resp *http.Response, err error) (bool, error) {
		if resp == nil {
			return false, err
		}
		switch resp.StatusCode {
		case http.StatusTooManyRequests, http.StatusBadGateway:
			return true, err
		case http.StatusServiceUnavailable:
			return !isLightningGatewayTimeout(resp), err
		}
		return false, err
	}
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return client.StandardClient()
}

// isLightningGatewayTimeout reports whether a 503 response is the API gateway's time limit rather than
// a temporary outage. It reads the body to find out and puts it back, so the response can still be read.
func isLightningGatewayTimeout(resp *http.Response) bool {
	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
	if err != nil {
		return false
	}

	var failure rtcesqlv1.Failure
	if err := json.Unmarshal(body, &failure); err != nil {
		return false
	}
	for _, e := range failure.GetErrors() {
		if e.GetCode() == lightningGatewayTimeoutCode {
			return true
		}
	}
	return false
}

// GetLightningUrl returns the regional endpoint that serves Lightning Tables queries for a Kafka cluster
// in the given cloud and region. There is no API to look it up, so it is derived from the platform
// hostname, like the public Flink endpoint: https://sql.<region>.<cloud>.<hostname>.
func GetLightningUrl(cfg *config.Config, cloud, region string) (string, error) {
	if cfg.IsTest {
		return testserver.TestLightningUrl.String(), nil
	}

	u, err := url.Parse(cfg.Context().GetPlatformServer())
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("https://sql.%s.%s.%s", region, strings.ToLower(cloud), u.Host), nil
}

func (c *LightningClient) lightningApiContext() context.Context {
	return context.WithValue(context.Background(), rtcesqlv1.ContextAccessToken, c.authToken)
}

// ExecuteQuery runs a SQL query against the Lightning Tables of a Kafka cluster and returns its inline
// result. A positive maxResultRows asks the server to return at most that many rows.
func (c *LightningClient) ExecuteQuery(environmentId, kafkaClusterId, sql string, maxResultRows int32) (rtcesqlv1.QueryV1QueryResult, error) {
	request := rtcesqlv1.NewQueryV1QueryRequest(environmentId, kafkaClusterId, sql)
	if maxResultRows > 0 {
		request.SetOptions(rtcesqlv1.QueryV1QueryOptions{MaxResultRows: rtcesqlv1.PtrInt32(maxResultRows)})
	}

	resp, httpResp, err := c.StatementsQueryV1Api.ExecuteQueryV1Statement(c.lightningApiContext()).QueryV1QueryRequest(*request).Execute()

	// lightningSyncOnlyHeader should rule out an asynchronous job. Check for one before anything else:
	// the SDK decodes a 202 as an inline result with no rows, which must not be reported as an empty result.
	if httpResp != nil && httpResp.StatusCode == http.StatusAccepted {
		return rtcesqlv1.QueryV1QueryResult{}, newLightningAsyncRequiredError()
	}

	if err != nil {
		return rtcesqlv1.QueryV1QueryResult{}, catchLightningError(err, httpResp)
	}

	return resp.GetResult(), nil
}

func catchLightningError(err error, httpResp *http.Response) error {
	if httpResp == nil {
		// The endpoint is derived rather than looked up, so a cloud or region without Lightning Tables
		// has no endpoint at all.
		var dnsErr *net.DNSError
		if goerrors.As(err, &dnsErr) && dnsErr.IsNotFound {
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(`no Lightning Tables endpoint found at "%s"`, dnsErr.Name),
				"Lightning Tables may not be available for the Kafka cluster's cloud provider or region.",
			)
		}
		return err
	}

	if httpResp.StatusCode < http.StatusMultipleChoices {
		return fmt.Errorf("failed to read the query result: %w", err)
	}

	if isLightningAsyncRequired(err, httpResp) {
		return newLightningAsyncRequiredError()
	}

	err = errors.CatchCCloudV2Error(err, httpResp)
	if _, ok := err.(rtcesqlv1.GenericOpenAPIError); ok {
		// Nothing could be read from the body (for example, an HTML page from a proxy), and the SDK's
		// error may only say that the body couldn't be decoded, so report the status instead.
		err = errors.New(httpResp.Status)
	}
	if requestId := httpResp.Header.Get("X-Request-Id"); requestId != "" && httpResp.StatusCode >= http.StatusInternalServerError {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(`When contacting Confluent Support, include request ID "%s".`, requestId))
	}
	return err
}

// isLightningAsyncRequired reports whether the query was rejected because answering it would require
// asynchronous execution, which lightningSyncOnlyHeader disallows.
func isLightningAsyncRequired(err error, httpResp *http.Response) bool {
	if httpResp == nil || httpResp.StatusCode != http.StatusUnprocessableEntity {
		return false
	}

	openApiError, ok := err.(rtcesqlv1.GenericOpenAPIError)
	if !ok {
		return false
	}

	var failure rtcesqlv1.Failure
	if err := json.Unmarshal(openApiError.Body(), &failure); err != nil {
		return false
	}

	for _, e := range failure.GetErrors() {
		if e.GetCode() == lightningAsyncRequiredCode {
			return true
		}
	}
	return false
}

func newLightningAsyncRequiredError() error {
	return errors.NewErrorWithSuggestions(
		"the query result is too large or too slow to return synchronously",
		`This command only runs queries that return their results synchronously. Narrow the query, for example with a "LIMIT" clause or a more selective "WHERE" clause.`,
	)
}
