package ccloudv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

// This file serves the unified `confluent logs` commands. It sits next to logging.go, which serves `confluent connect
// logs`, and reuses its logging host lookup without changing it.

// getLoggingServerUrl returns the logging API's server URL, without a path. The host is chosen by getLoggingUrl, so the
// mapping from a platform to its logging host stays in one place.
func getLoggingServerUrl(baseURL string) (string, error) {
	searchUrl, err := getLoggingUrl(baseURL)
	if err != nil {
		return "", err
	}

	u, err := url.Parse(searchUrl)
	if err != nil {
		return "", fmt.Errorf("failed to parse logging URL: %w", err)
	}
	u.Path = ""
	return u.String(), nil
}

// newLoggingClient builds the SDK client for the logging API, which is served from its own host rather than the
// regular Confluent Cloud API URL, and a context carrying the data plane token it authenticates with.
func (c *Client) newLoggingClient() (*loggingv1.APIClient, context.Context, error) {
	loggingURL, err := getLoggingServerUrl(c.cfg.Context().GetPlatformServer())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get logging API URL: %w", err)
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	cfg := loggingv1.NewConfiguration()
	cfg.HTTPClient = newRetryableHttpClientReturningErrors(c.cfg, false)
	cfg.Servers = loggingv1.ServerConfigurations{{URL: loggingURL}}
	cfg.UserAgent = c.cfg.Version.UserAgent

	ctx := context.WithValue(context.Background(), loggingv1.ContextAccessToken, dataplaneToken)
	return loggingv1.NewAPIClient(cfg), ctx, nil
}

// ListLogSources queries GET /logs/v1/sources.
func (c *Client) ListLogSources() (*loggingv1.LoggingV1SourceList, error) {
	client, ctx, err := c.newLoggingClient()
	if err != nil {
		return nil, err
	}

	sources, httpResp, err := client.LogsLoggingV1Api.ListLoggingV1Sources(ctx).Execute()
	if err != nil {
		return nil, catchLoggingError(err, httpResp)
	}
	return &sources, nil
}

// DescribeLogSource queries GET /logs/v1/sources/{name}.
func (c *Client) DescribeLogSource(name string) (*loggingv1.LoggingV1Source, error) {
	client, ctx, err := c.newLoggingClient()
	if err != nil {
		return nil, err
	}

	source, httpResp, err := client.LogsLoggingV1Api.DescribeLoggingV1Source(ctx, name).Execute()
	if err != nil {
		return nil, catchLoggingError(err, httpResp)
	}
	return &source, nil
}

// catchLoggingError turns a logging API failure into a CLI error. A 429 tells the user how long to wait, from the
// Retry-After header the API sends.
func catchLoggingError(err error, httpResp *http.Response) error {
	if httpResp != nil && httpResp.StatusCode == http.StatusTooManyRequests {
		return rateLimitError(httpResp.Header.Get("Retry-After"), time.Now())
	}
	return errors.CatchCCloudV2Error(err, httpResp)
}

// rateLimitError reports a 429. Retry-After is either a number of seconds or an HTTP date.
func rateLimitError(retryAfter string, now time.Time) error {
	const message = "rate limit exceeded for log queries"

	if seconds, err := strconv.Atoi(strings.TrimSpace(retryAfter)); err == nil && seconds > 0 {
		return fmt.Errorf("%s. Retry after %ds", message, seconds)
	}
	if date, err := http.ParseTime(retryAfter); err == nil {
		if wait := date.Sub(now); wait > 0 {
			return fmt.Errorf("%s. Retry after %ds", message, int((wait+time.Second-1)/time.Second))
		}
	}
	return fmt.Errorf("%s. Try again later", message)
}

// newLogsSearchRequest builds POST /logs/v1/search for a unified query. It is the counterpart of getLoggingRequest in
// logging.go, which only takes the request type of `confluent connect logs`.
func newLogsSearchRequest(loggingURL string, request LogsSearchRequest, dataplaneToken string, pageSize int, pageToken string) (*http.Request, error) {
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	u, err := url.Parse(loggingURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse logging URL: %w", err)
	}
	query := u.Query()
	if pageSize > 0 {
		query.Set("page_size", strconv.Itoa(pageSize))
	}
	if pageToken != "" {
		query.Set("page_token", pageToken)
	}
	u.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+dataplaneToken)
	return req, nil
}

// LogsSearchRequest is the body of POST /logs/v1/search for the unified `confluent logs` commands. Unlike
// LoggingSearchRequest, it carries the log source and an open-ended search map.
type LogsSearchRequest struct {
	CRN       string         `json:"crn"`
	Source    string         `json:"source,omitempty"`
	Search    map[string]any `json:"search,omitempty"`
	Sort      string         `json:"sort,omitempty"`
	StartTime string         `json:"start_time"`
	EndTime   string         `json:"end_time"`
}

// LogsSearchResponse holds log records as generic maps, because each log source declares its own returned fields.
type LogsSearchResponse struct {
	Data       []map[string]any `json:"data"`
	Metadata   *LoggingMetadata `json:"metadata,omitempty"`
	ApiVersion string           `json:"api_version"`
	Kind       string           `json:"kind"`
}

type loggingErrorResponse struct {
	Errors []struct {
		Status string `json:"status"`
		Detail string `json:"detail"`
	} `json:"errors"`
}

// SearchLogs queries POST /logs/v1/search for any log source.
func (c *Client) SearchLogs(request LogsSearchRequest, pageSize int, pageToken string) (*LogsSearchResponse, error) {
	loggingURL, err := getLoggingUrl(c.cfg.Context().GetPlatformServer())
	if err != nil {
		return nil, fmt.Errorf("failed to get logging API URL: %w", err)
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	req, err := newLogsSearchRequest(loggingURL, request, dataplaneToken, pageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging API request: %w", err)
	}

	body, err := c.doLoggingRequest(req)
	if err != nil {
		return nil, err
	}

	return decodeLogsSearchResponse(body)
}

func (c *Client) doLoggingRequest(req *http.Request) ([]byte, error) {
	resp, err := newRetryableHttpClientReturningErrors(c.cfg, false).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to logging API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("logging API request failed with status %d and failed to read response body: %w", resp.StatusCode, err)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, rateLimitError(resp.Header.Get("Retry-After"), time.Now())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, loggingStatusError(resp.StatusCode, body)
	}
	return body, nil
}

// loggingStatusError surfaces the API's own error detail, for example "Credentials are required to access this resource.".
func loggingStatusError(statusCode int, body []byte) error {
	var errResp loggingErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 && errResp.Errors[0].Detail != "" {
		return fmt.Errorf("%s", errResp.Errors[0].Detail)
	}
	return fmt.Errorf("logging API request failed with status %d", statusCode)
}

// decodeLogsSearchResponse decodes JSON numbers without going through float64, so integers above 2^53 survive into
// the JSON, YAML and JSONL output. Numbers are returned as int64, uint64 or float64 rather than json.Number, because
// the YAML encoder would print a json.Number as a quoted string.
func decodeLogsSearchResponse(body []byte) (*LogsSearchResponse, error) {
	var response LogsSearchResponse
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	if err := decoder.Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode logging API response: %w", err)
	}
	for _, record := range response.Data {
		for key, value := range record {
			record[key] = convertJsonNumbers(value)
		}
	}
	return &response, nil
}

func convertJsonNumbers(value any) any {
	switch v := value.(type) {
	case json.Number:
		if i, err := strconv.ParseInt(v.String(), 10, 64); err == nil {
			return i
		}
		if u, err := strconv.ParseUint(v.String(), 10, 64); err == nil {
			return u
		}
		if f, err := v.Float64(); err == nil {
			return f
		}
		return v.String()
	case map[string]any:
		for key, item := range v {
			v[key] = convertJsonNumbers(item)
		}
	case []any:
		for i, item := range v {
			v[i] = convertJsonNumbers(item)
		}
	}
	return value
}
