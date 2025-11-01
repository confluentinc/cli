package ccloudv2

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	UsageLimitsPath = "/usage_limits"
)

type UsageLimitValue struct {
	Unit      string `json:"unit"`
	Value     int32  `json:"value"`
	Unlimited bool   `json:"unlimited,omitempty"`
}

type Limits struct {
	Ingress *UsageLimitValue `json:"ingress,omitempty"`
	Egress  *UsageLimitValue `json:"egress,omitempty"`
	Storage *UsageLimitValue `json:"storage,omitempty"`
	MaxEcku *UsageLimitValue `json:"max_ecku,omitempty"`
}

type TierLimit struct {
	ClusterLimits Limits `json:"cluster_limits"`
}

type UsageLimits struct {
	TierLimits map[string]TierLimit `json:"tier_limits"`
	CkuLimits  map[string]Limits    `json:"cku_limits"`
}

type UsageLimitsResponse struct {
	UsageLimits UsageLimits `json:"usage_limits"`
	Error       *string     `json:"error,omitempty"`
}

func (c *Limits) GetIngress() int32 {
	if c == nil || c.Ingress == nil {
		return 0
	}
	return c.Ingress.Value
}

func (c *Limits) GetEgress() int32 {
	if c == nil || c.Egress == nil {
		return 0
	}
	return c.Egress.Value
}

func (c *Limits) GetStorage() *UsageLimitValue {
	if c == nil {
		return nil
	}
	return c.Storage
}

func (c *Limits) GetMaxEcku() *UsageLimitValue {
	if c == nil {
		return nil
	}
	return c.MaxEcku
}

func (t *TierLimit) GetClusterLimits() *Limits {
	if t == nil {
		return nil
	}
	return &t.ClusterLimits
}

func (u *UsageLimits) GetCkuLimit(cku int32) *Limits {
	if u == nil {
		return nil
	}
	ckuStr := strconv.FormatInt(int64(cku), 10)
	ckuLimit, ok := u.CkuLimits[ckuStr]
	if !ok {
		return nil
	}
	return &ckuLimit
}

func (u *UsageLimits) GetTierLimit(sku string) *TierLimit {
	if u == nil {
		return nil
	}
	tierLimit, ok := u.TierLimits[sku]
	if !ok {
		return nil
	}
	return &tierLimit
}

func (c *Client) GetUsageLimits(provider, lkcId, envId string) (*UsageLimits, error) {
	baseURL := getServerUrl(c.cfg.Context().GetPlatformServer())
	usageLimitsURL, err := getUsageLimitsUrl(baseURL, provider, lkcId, envId)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage limits API URL: %w", err)
	}

	authToken := c.cfg.Context().GetAuthToken()
	if authToken == "" {
		return nil, fmt.Errorf("failed to get auth token")
	}

	req, err := getUsageLimitsRequest(usageLimitsURL, authToken)
	if err != nil {
		return nil, fmt.Errorf("failed to create usage limits API request: %w", err)
	}

	httpClient := NewRetryableHttpClient(c.cfg, false)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make HTTP request to usage limits API due to issue connecting to the server: %w", err)
	}
	defer resp.Body.Close()

	responseBody, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, fmt.Errorf("usage limits API request failed with status %d and failed to read response body: %w", resp.StatusCode, readErr)
	}

	if resp.StatusCode != http.StatusOK {
		if len(responseBody) > 0 {
			return nil, fmt.Errorf("usage limits API request failed with status %d\nResponse body: %s", resp.StatusCode, string(responseBody))
		}
		return nil, fmt.Errorf("usage limits API request failed with status %d", resp.StatusCode)
	}

	var usageLimitsResponse UsageLimitsResponse
	if err := json.Unmarshal(responseBody, &usageLimitsResponse); err != nil {
		return nil, fmt.Errorf("failed to decode usage limits API response: %w\nResponse body: %s\nResponse status: %d", err, string(responseBody), resp.StatusCode)
	}

	if usageLimitsResponse.Error != nil {
		return nil, fmt.Errorf("usage limits API request failed: %s", *usageLimitsResponse.Error)
	}

	return &usageLimitsResponse.UsageLimits, nil
}

func getUsageLimitsRequest(usageLimitsURL, authToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, usageLimitsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	return req, nil
}

func getUsageLimitsUrl(serverURL, provider, lkcId, envId string) (string, error) {
	// Normalize to API server using shared util, which ensures:
	// - confluent/cloud dev/stage -> api.<host> with empty path
	// - local hosts -> path "/api"
	u, err := url.Parse(serverURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse server URL: %w", err)
	}

	// Append usage_limits to existing path
	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" {
		u.Path = UsageLimitsPath
	} else {
		u.Path = "/" + strings.TrimLeft(basePath, "/") + UsageLimitsPath
	}

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

	return u.String(), nil
}
