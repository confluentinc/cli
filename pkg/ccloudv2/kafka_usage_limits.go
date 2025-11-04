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

	ErrMsgUsageLimitsAPI      = "usage limits API HTTP request failed"
	ErrFailedToGetAuthToken   = "failed to get auth token"
	ErrFailedToDecodeResponse = "failed to decode response"
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
	usageLimitsURL := c.getUsageLimitsUrl(provider, lkcId, envId)

	authToken := c.cfg.Context().GetAuthToken()
	if authToken == "" {
		return nil, usageLimitsError(ErrFailedToGetAuthToken)
	}

	req, err := getUsageLimitsRequest(usageLimitsURL, authToken)
	if err != nil {
		return nil, usageLimitsError(err)
	}

	httpClient := NewRetryableHttpClient(c.cfg, false)
	resp, err := httpClient.Do(req)
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
		return nil, usageLimitsError(fmt.Errorf("%s: %w", ErrFailedToDecodeResponse, err))
	}

	if usageLimitsResponse.Error != nil {
		return nil, usageLimitsError(*usageLimitsResponse.Error)
	}

	return &usageLimitsResponse.UsageLimits, nil
}

func usageLimitsError(msg interface{}) error {
	if err, ok := msg.(error); ok {
		return fmt.Errorf("%s: %w", ErrMsgUsageLimitsAPI, err)
	}
	return fmt.Errorf("%s: %+v", ErrMsgUsageLimitsAPI, msg)
}

func getUsageLimitsRequest(usageLimitsURL, authToken string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, usageLimitsURL, nil)
	if err != nil {
		return nil, usageLimitsError(err)
	}
	req.Header.Set("Authorization", "Bearer "+authToken)
	return req, nil
}

func (c *Client) getUsageLimitsUrl(provider, lkcId, envId string) string {
	serverURL := getServerUrl(c.cfg.Context().GetPlatformServer())
	u, err := url.Parse(serverURL)
	if err != nil {
		return serverURL
	}

	// Append usage_limits to existing path
	basePath := strings.TrimRight(u.Path, "/")
	if basePath == "" {
		u.Path = UsageLimitsPath
	} else {
		// local testing hosts
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

	return u.String()
}
