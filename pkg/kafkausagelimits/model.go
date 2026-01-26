package kafkausagelimits

import (
	"net/http"
	"strconv"

	"github.com/confluentinc/cli/v4/pkg/config"
)

type UsageLimitsClient struct {
	HttpClient *http.Client
	cfg        *config.Config
}

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
