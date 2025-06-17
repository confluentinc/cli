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
	"github.com/confluentinc/cli/v4/pkg/errors"
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
}

type LoggingLogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	TaskId    string            `json:"task_id,omitempty"`
	Id        string            `json:"id,omitempty"`
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

// SearchConnectorLogs searches logs for a specific connector using the Logging API
func (c *Client) SearchConnectorLogs(environmentId, kafkaClusterId, connectorId, startTime, endTime string, levels []string, searchText string, pageSize int, pageToken string) (*LoggingSearchResponse, error) {
	// Build the CRN for the connector
	crn := fmt.Sprintf("crn://confluent.cloud/organization=%s/environment=%s/cloud-cluster=%s/connector=%s",
		c.cfg.Context().GetCurrentOrganization(),
		environmentId,
		kafkaClusterId,
		connectorId,
	)

	// Build the logging API URL using the specific logging subdomain
	baseURL := c.cfg.Context().GetPlatformServer()
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse platform server URL: %w", err)
	}

	// Use the logging-specific subdomain
	if u.Host == "127.0.0.1:1024" || u.Host == "localhost:1024" {
		// In test/mock environments, do not rewrite the host
	} else if strings.Contains(u.Host, "devel.cpdev.cloud") {
		u.Host = "api.logging.devel.cpdev.cloud"
	} else if strings.Contains(u.Host, "stag.cpdev.cloud") {
		u.Host = "api.logging.stag.cpdev.cloud"
	} else if strings.Contains(u.Host, "confluent.cloud") {
		u.Host = "api.logging.confluent.cloud"
	} else {
		// Fallback for other environments
		u.Host = "api.logging." + strings.TrimPrefix(u.Host, "api.")
	}
	u.Path = "/logs/v1/search"
	loggingURL := u.String()

	// Build the request body
	request := LoggingSearchRequest{
		CRN: crn,
		Search: LoggingSearchParams{
			Level:      levels,
			SearchText: searchText,
		},
		Sort:      "desc",
		StartTime: startTime,
		EndTime:   endTime,
	}

	// Marshal the request body
	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Build the URL with query parameters
	u, err = url.Parse(loggingURL)
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

	// Create the HTTP request
	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, u.String(), bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Get data plane token instead of control plane token
	dataplaneToken, err := auth.GetDataplaneToken(c.cfg.Context())
	if err != nil {
		return nil, fmt.Errorf("failed to get data plane token: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+dataplaneToken)

	// Get HTTP client using the same pattern as other clients
	httpClient := NewRetryableHttpClient(c.cfg, false)

	// Make the HTTP request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to logging API: %w\nRequest URL: %s\nRequest method: %s", err, u.String(), req.Method)
	}
	defer resp.Body.Close()

	// Read response body for error details
	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("logging API request failed with status %d and failed to read response body: %w", resp.StatusCode, readErr)
	}

	// Check for HTTP errors with detailed information
	if resp.StatusCode != http.StatusOK {
		errorMsg := fmt.Sprintf("logging API request failed with status %d", resp.StatusCode)

		// Add response body if available
		if len(responseBody) > 0 {
			errorMsg += fmt.Sprintf("\nResponse body: %s", string(responseBody))
		}

		// Add request details for debugging
		errorMsg += fmt.Sprintf("\nRequest URL: %s", u.String())
		errorMsg += fmt.Sprintf("\nRequest CRN: %s", crn)
		errorMsg += fmt.Sprintf("\nRequest time range: %s to %s", startTime, endTime)

		// Add specific suggestions based on status code
		var suggestions string
		switch resp.StatusCode {
		case http.StatusUnauthorized:
			suggestions = "Please check your authentication. Try running 'confluent login' to refresh your credentials."
		case http.StatusForbidden:
			suggestions = "You don't have permission to access logs for this connector. Please check your RBAC permissions."
		case http.StatusNotFound:
			suggestions = "The connector or resource was not found. Please verify the connector ID, environment, and cluster are correct."
		case http.StatusBadRequest:
			suggestions = "The request parameters are invalid. Please check your time format (RFC3339), connector ID, and other parameters."
		case http.StatusTooManyRequests:
			suggestions = "Rate limit exceeded. Please wait a moment before retrying."
		case http.StatusInternalServerError:
			suggestions = "Internal server error occurred. Please try again later or contact support if the issue persists."
		default:
			suggestions = "Please check your connector ID, environment, cluster settings, and ensure you have proper permissions to access logs."
		}

		return nil, errors.NewErrorWithSuggestions(errorMsg, suggestions)
	}

	// Parse the response (recreate reader since we already read the body)
	var logsResponse LoggingSearchResponse
	if err := json.Unmarshal(responseBody, &logsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode logging API response: %w\nResponse body: %s\nResponse status: %d", err, string(responseBody), resp.StatusCode)
	}

	return &logsResponse, nil
}
