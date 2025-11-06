package kafkausagelimits

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/confluentinc/cli/v4/pkg/config"
)

const (
	UsageLimitsPath = "/api/usage_limits"

	UsageLimitsAPIErrorMsg         = "usage limits API HTTP request failed"
	FailedToGetAuthTokenErrorMsg   = "failed to get auth token"
	FailedToDecodeResponseErrorMsg = "failed to decode response"
)

func NewUsageLimitsClient(cfg *config.Config, httpClient *http.Client) *UsageLimitsClient {
	return &UsageLimitsClient{
		HttpClient: httpClient,
		cfg:        cfg,
	}
}

func (c *UsageLimitsClient) GetUsageLimits(provider, lkcId, envId string) (*UsageLimits, error) {
	usageLimitsURL := c.getUsageLimitsUrl(provider, lkcId, envId)

	authToken := c.cfg.Context().GetAuthToken()
	if authToken == "" {
		return nil, usageLimitsError(FailedToGetAuthTokenErrorMsg)
	}

	req, err := getUsageLimitsRequest(usageLimitsURL, authToken)
	if err != nil {
		return nil, usageLimitsError(err)
	}

	resp, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, usageLimitsError(err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, usageLimitsError(err)
	}

	if resp.StatusCode != http.StatusOK {
		if len(responseBody) > 0 {
			return nil, usageLimitsError(fmt.Sprintf("status %d\nResponse body: %s", resp.StatusCode, string(responseBody)))
		}
		return nil, usageLimitsError(fmt.Sprintf("status %d", resp.StatusCode))
	}

	var usageLimitsResponse UsageLimitsResponse
	if err = json.Unmarshal(responseBody, &usageLimitsResponse); err != nil {
		return nil, usageLimitsError(fmt.Errorf("%s: %w", FailedToDecodeResponseErrorMsg, err))
	}

	if usageLimitsResponse.Error != nil {
		return nil, usageLimitsError(*usageLimitsResponse.Error)
	}

	return &usageLimitsResponse.UsageLimits, nil
}

func usageLimitsError(msg interface{}) error {
	if err, ok := msg.(error); ok {
		return fmt.Errorf("%s: %w", UsageLimitsAPIErrorMsg, err)
	}
	return fmt.Errorf("%s: %+v", UsageLimitsAPIErrorMsg, msg)
}

func getUsageLimitsRequest(usageLimitsURL, authToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, usageLimitsURL, nil)
	if err != nil {
		return nil, usageLimitsError(err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	return req, nil
}

func (c *UsageLimitsClient) getUsageLimitsUrl(provider, lkcId, envId string) string {
	serverURL := c.cfg.Context().GetPlatformServer()
	u, err := url.Parse(serverURL)
	if err != nil {
		return serverURL
	}

	u.Path = UsageLimitsPath

	// Build query parameters
	query := u.Query()
	if provider != "" {
		query.Set("provider", provider)
	}
	if lkcId != "" {
		query.Set("lkc_id", lkcId)
	}
	if envId != "" {
		query.Set("env_id", envId)
	}
	u.RawQuery = query.Encode()

	return u.String()
}
