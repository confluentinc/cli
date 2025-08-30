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

	"github.com/confluentinc/cli/v4/pkg/auth"
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

func getLoggingRequest(loggingURL string, request LoggingSearchRequest, dataplaneToken string, pageSize int, pageToken string) (*http.Request, error) {
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
	} else {
		u.Host = "api.logging." + strings.TrimPrefix(u.Host, "api.")
	}
	u.Path = "/logs/v1/search"
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
