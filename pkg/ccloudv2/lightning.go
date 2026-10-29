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
	"time"

	"github.com/hashicorp/go-retryablehttp"

	rtcesqlv1 "github.com/confluentinc/ccloud-sdk-go-v2/rtce-sql/v1"

	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	plog "github.com/confluentinc/cli/v4/pkg/log"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

const (
	// lightningSyncOnlyHeader makes the server reject a query it can't answer synchronously instead of starting an async job.
	lightningSyncOnlyHeader = "X-Lightning-Sync-Only"

	lightningAsyncRequiredCode = "FE_ASYNC_REQUIRED"

	// lightningGatewayTimeoutCode marks an HTTP 503 caused by the API gateway's time limit.
	lightningGatewayTimeoutCode = "service_timeout"

	// lightningQueryTimeout bounds a query and its retries; the API gateway itself gives up after 60 seconds.
	lightningQueryTimeout = 2 * time.Minute

	lightningNarrowQuerySuggestion = `Narrow the query, for example with a "LIMIT" clause or a more selective "WHERE" clause.`
)

type LightningClient struct {
	*rtcesqlv1.APIClient
	authToken    string
	queryTimeout time.Duration
}

func NewLightningClient(url, userAgent string, unsafeTrace bool, authToken string) *LightningClient {
	cfg := rtcesqlv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.DefaultHeader = map[string]string{lightningSyncOnlyHeader: "true"}
	cfg.HTTPClient = newLightningHttpClient(unsafeTrace)
	cfg.Servers = rtcesqlv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return &LightningClient{
		APIClient:    rtcesqlv1.NewAPIClient(cfg),
		authToken:    authToken,
		queryTimeout: lightningQueryTimeout,
	}
}

// newLightningHttpClient retries only throttling and gateway errors; a 500 or a gateway timeout would likely fail again.
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
	// Return the last response, so the user sees the server's message.
	client.ErrorHandler = retryablehttp.PassthroughErrorHandler
	return client.StandardClient()
}

// isLightningGatewayTimeout reports whether a 503 is the API gateway's time limit, leaving the body readable.
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

// GetLightningUrl derives the regional endpoint https://sql.<region>.<cloud>.<hostname>, since no API returns it.
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

// ExecuteQuery runs a SQL query; a positive maxResultRows limits how many rows the server returns.
func (c *LightningClient) ExecuteQuery(environmentId, kafkaClusterId, sql string, maxResultRows int32) (rtcesqlv1.QueryV1QueryResult, error) {
	request := rtcesqlv1.NewQueryV1QueryRequest(environmentId, kafkaClusterId, sql)
	if maxResultRows > 0 {
		request.SetOptions(rtcesqlv1.QueryV1QueryOptions{MaxResultRows: rtcesqlv1.PtrInt32(maxResultRows)})
	}

	ctx, cancel := context.WithTimeout(c.lightningApiContext(), c.queryTimeout)
	defer cancel()

	resp, httpResp, err := c.StatementsQueryV1Api.ExecuteQueryV1Statement(ctx).QueryV1QueryRequest(*request).Execute()

	// The SDK decodes an async job (HTTP 202) as an empty result, so check for one first.
	if httpResp != nil && httpResp.StatusCode == http.StatusAccepted {
		return rtcesqlv1.QueryV1QueryResult{}, newLightningAsyncRequiredError()
	}

	if err != nil {
		if goerrors.Is(ctx.Err(), context.DeadlineExceeded) {
			return rtcesqlv1.QueryV1QueryResult{}, errors.NewErrorWithSuggestions(
				"the query timed out waiting for a response",
				"Check your network connection and proxy settings, then try again.",
			)
		}
		return rtcesqlv1.QueryV1QueryResult{}, catchLightningError(err, httpResp)
	}

	return resp.GetResult(), nil
}

func catchLightningError(err error, httpResp *http.Response) error {
	if httpResp == nil {
		// The endpoint is derived, so a cloud or region without Lightning Tables has no DNS record.
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

	// The SQL is the only thing in the request body.
	if httpResp.StatusCode == http.StatusRequestEntityTooLarge {
		return errors.NewErrorWithSuggestions("the SQL statement is too large", "Shorten the statement and try again.")
	}

	// The API gateway's time limit, which the same query would likely hit again.
	if httpResp.StatusCode == http.StatusGatewayTimeout || (httpResp.StatusCode == http.StatusServiceUnavailable && isLightningGatewayTimeout(httpResp)) {
		return errors.NewErrorWithSuggestions("the query took too long to complete", lightningNarrowQuerySuggestion)
	}

	err = errors.CatchCCloudV2Error(err, httpResp)
	if _, ok := err.(rtcesqlv1.GenericOpenAPIError); ok {
		// The body couldn't be read, for example a proxy's HTML page, so report the status.
		err = errors.New(httpResp.Status)
	}
	if requestId := httpResp.Header.Get("X-Request-Id"); requestId != "" && httpResp.StatusCode >= http.StatusInternalServerError {
		return errors.NewErrorWithSuggestions(err.Error(), fmt.Sprintf(`When contacting Confluent Support, include request ID "%s".`, requestId))
	}
	return err
}

// isLightningAsyncRequired reports whether the server rejected a query that needs async execution.
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
		"This command only runs queries that return their results synchronously. "+lightningNarrowQuerySuggestion,
	)
}
