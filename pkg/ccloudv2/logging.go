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

	loggingv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/logging/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

const (
	LogsPageSize = 200
)

type LoggingSearchRequest struct {
	CRN       string              `json:"crn"`
	Search    LoggingSearchParams `json:"search"`
	Sort      string              `json:"sort"`
	StartTime string              `json:"start_time"`
	EndTime   string              `json:"end_time"`
}

type LoggingSearchParams struct {
	Level      []string `json:"level,omitempty"`
	SearchText string   `json:"search_text,omitempty"`
	Id         string   `json:"id,omitempty"`
}

type LoggingLogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	TaskId    string            `json:"task_id,omitempty"`
	Exception *LoggingException `json:"exception,omitempty"`
}

type LoggingException struct {
	Stacktrace string `json:"stacktrace,omitempty"`
}

type LoggingMetadata struct {
	Next string `json:"next,omitempty"`
}

type LoggingSearchResponse struct {
	Data       []LoggingLogEntry `json:"data"`
	Metadata   *LoggingMetadata  `json:"metadata,omitempty"`
	ApiVersion string            `json:"api_version"`
	Kind       string            `json:"kind"`
}

func (c *Client) SearchConnectorLogs(crn, connectorId, startTime, endTime string, levels []string, searchText string, pageToken string) (*LoggingSearchResponse, error) {
	baseURL := c.cfg.Context().GetPlatformServer()
	loggingURL, err := getLoggingUrl(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get logging API URL: %w", err)
	}

	request := LoggingSearchRequest{
		CRN: crn,
		Search: LoggingSearchParams{
			Level:      levels,
			SearchText: searchText,
			Id:         connectorId,
		},
		Sort:      "desc",
		StartTime: startTime,
		EndTime:   endTime,
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	req, err := getLoggingRequest(loggingURL, request, dataplaneToken, LogsPageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging API request: %w", err)
	}

	httpClient := NewRetryableHttpClient(c.cfg, false)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to logging API due to issue connecting to the server, please make sure you adhere to 5 requests per minute per connector rate limit")
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("logging API request failed with status %d and failed to read response body: %w", resp.StatusCode, readErr)
	}

	err = getResponseErrorMessage(resp.StatusCode, responseBody, crn)
	if err != nil {
		return nil, err
	}

	var logsResponse LoggingSearchResponse
	if err := json.Unmarshal(responseBody, &logsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode logging API response: %w\nResponse body: %s\nResponse status: %d", err, string(responseBody), resp.StatusCode)
	}

	return &logsResponse, nil
}

func getLoggingRequest(loggingURL string, request any, dataplaneToken string, pageSize int, pageToken string) (*http.Request, error) {
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

func getLoggingUrl(baseURL string) (string, error) {
	return getLoggingUrlForPath(baseURL, "/logs/v1/search")
}

func getLoggingUrlForPath(baseURL, path string) (string, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse platform server URL: %w", err)
	}

	if u.Host == "127.0.0.1:1024" || u.Host == "localhost:1024" {

	} else if strings.Contains(u.Host, "devel.cpdev.cloud") {
		u.Host = "api.logging.devel.cpdev.cloud"
	} else if strings.Contains(u.Host, "stag.cpdev.cloud") {
		u.Host = "api.logging.stag.cpdev.cloud"
	} else if strings.Contains(u.Host, "confluent.cloud") {
		u.Host = "api.logging.confluent.cloud"
	} else if strings.Contains(u.Host, "dr-test.cpdev.cloud") {
		u.Host = "api.logging.dr-test.cpdev.cloud"
	} else {
		u.Host = "api.logging." + strings.TrimPrefix(u.Host, "api.")
	}
	u.Path = path
	loggingURL := u.String()
	return loggingURL, nil
}

func getResponseErrorMessage(statusCode int, responseBody []byte, crn string) error {
	if statusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("logging API request failed with status %d", statusCode)

		if len(responseBody) > 0 {
			errorMsg += fmt.Sprintf("\nResponse body: %s", string(responseBody))
		}
		errorMsg += fmt.Sprintf("\nRequest CRN: %s", crn)
		return fmt.Errorf("%s\n", errorMsg)
	}
	return nil
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

	req, err := getLoggingRequest(loggingURL, request, dataplaneToken, pageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create logging API request: %w", err)
	}

	body, err := c.doLoggingRequest(req)
	if err != nil {
		return nil, err
	}

	var response LogsSearchResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to decode logging API response: %w", err)
	}
	return &response, nil
}

// newLoggingClient builds the SDK client for the logging API, which is served from its own host rather than the
// regular Confluent Cloud API URL, and a context carrying the data plane token it authenticates with.
func (c *Client) newLoggingClient() (*loggingv1.APIClient, context.Context, error) {
	loggingURL, err := getLoggingUrlForPath(c.cfg.Context().GetPlatformServer(), "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get logging API URL: %w", err)
	}

	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	cfg := loggingv1.NewConfiguration()
	cfg.HTTPClient = NewRetryableHttpClient(c.cfg, false)
	cfg.Servers = loggingv1.ServerConfigurations{{URL: loggingURL}}
	cfg.UserAgent = c.cfg.Version.UserAgent

	ctx := context.WithValue(context.Background(), loggingv1.ContextAccessToken, dataplaneToken)
	return loggingv1.NewAPIClient(cfg), ctx, nil
}

// DescribeLogSource queries GET /logs/v1/sources/{name}.
func (c *Client) DescribeLogSource(name string) (*loggingv1.LoggingV1Source, error) {
	client, ctx, err := c.newLoggingClient()
	if err != nil {
		return nil, err
	}

	source, httpResp, err := client.LogsLoggingV1Api.DescribeLoggingV1Source(ctx, name).Execute()
	if err != nil {
		return nil, errors.CatchCCloudV2Error(err, httpResp)
	}
	return &source, nil
}

func (c *Client) doLoggingRequest(req *http.Request) ([]byte, error) {
	httpClient := NewRetryableHttpClient(c.cfg, false)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to logging API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("logging API request failed with status %d and failed to read response body: %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, loggingStatusError(resp.StatusCode, body)
	}
	return body, nil
}

// loggingStatusError surfaces the API's own error detail, e.g. "Unknown source: connector".
func loggingStatusError(statusCode int, body []byte) error {
	var errResp loggingErrorResponse
	if err := json.Unmarshal(body, &errResp); err == nil && len(errResp.Errors) > 0 && errResp.Errors[0].Detail != "" {
		return fmt.Errorf("%s", errResp.Errors[0].Detail)
	}
	return fmt.Errorf("logging API request failed with status %d", statusCode)
}
