package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
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

// Handler for: "/api/usage_limits"
func handleUsageLimits(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")

		// Check for lkc_id query parameter to determine which limits to return
		lkcId := r.URL.Query().Get("lkc_id")
		switch lkcId {
		case "lkc-with-ecku-limits", "lkc-describe-with-ecku-limits":
			handleGetUsageLimitsEckuLimits(t)(w, r)
		case "lkc-with-usage-limits-error":
			handleGetUsageLimitsWithError(t)(w, r)
		default:
			handleGetUsageLimitsDefaultLimits(t)(w, r)
		}
	}
}

// Handler for GET "/usage_limits" - default limits
func handleGetUsageLimitsDefaultLimits(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limits := getDefaultUsageLimits()
		response := UsageLimitsResponse{
			UsageLimits: *limits,
			Error:       nil,
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}
}

// Handler for GET "/usage_limits" - Basic/Standard limits set with eCKU values
func handleGetUsageLimitsEckuLimits(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limits := getDefaultEckuUsageLimits()
		response := UsageLimitsResponse{
			UsageLimits: *limits,
			Error:       nil,
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}
}

func handleGetUsageLimitsWithError(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		errorMsg := "API error message"
		err := json.NewEncoder(w).Encode(UsageLimitsResponse{Error: &errorMsg})
		require.NoError(t, err)
	}
}

func getDefaultUsageLimits() *UsageLimits {
	return &UsageLimits{
		TierLimits: map[string]TierLimit{
			"BASIC": {
				ClusterLimits: Limits{
					Ingress: &UsageLimitValue{Unit: "MBPS", Value: 250},
					Egress:  &UsageLimitValue{Unit: "MBPS", Value: 750},
					Storage: &UsageLimitValue{Unit: "GB", Value: 5000},
				},
			},
			"STANDARD": {
				ClusterLimits: Limits{
					Ingress: &UsageLimitValue{Unit: "MBPS", Value: 250},
					Egress:  &UsageLimitValue{Unit: "MBPS", Value: 750},
					Storage: &UsageLimitValue{Unlimited: true},
				},
			},
			"ENTERPRISE": {
				ClusterLimits: Limits{
					Ingress: &UsageLimitValue{Unit: "MBPS", Value: 60},
					Egress:  &UsageLimitValue{Unit: "MBPS", Value: 180},
					Storage: &UsageLimitValue{Unlimited: true},
					MaxEcku: &UsageLimitValue{Value: 10},
				},
			},
			"FREIGHT": {
				ClusterLimits: Limits{
					Ingress: &UsageLimitValue{Unit: "MBPS", Value: 60},
					Egress:  &UsageLimitValue{Unit: "MBPS", Value: 180},
					Storage: &UsageLimitValue{Unlimited: true},
					MaxEcku: &UsageLimitValue{Value: 152},
				},
			},
		},
		CkuLimits: map[string]Limits{
			"1": {
				Ingress: &UsageLimitValue{Unit: "MBPS", Value: 60},
				Egress:  &UsageLimitValue{Unit: "MBPS", Value: 180},
				Storage: &UsageLimitValue{Unlimited: true},
			},
			"2": {
				Ingress: &UsageLimitValue{Unit: "MBPS", Value: 120},
				Egress:  &UsageLimitValue{Unit: "MBPS", Value: 360},
				Storage: &UsageLimitValue{Unlimited: true},
			},
			"3": {
				Ingress: &UsageLimitValue{Unit: "MBPS", Value: 180},
				Egress:  &UsageLimitValue{Unit: "MBPS", Value: 540},
				Storage: &UsageLimitValue{Unlimited: true},
			},
		},
	}
}

// Used for eCKU-specific usage limits
func getDefaultEckuUsageLimits() *UsageLimits {
	usageLimits := getDefaultUsageLimits()

	// Set Basic eCKU limits
	basicLimits := usageLimits.TierLimits["BASIC"]
	basicLimits.ClusterLimits.MaxEcku = &UsageLimitValue{Value: 50}
	basicLimits.ClusterLimits.Ingress = &UsageLimitValue{Unit: "MBPS", Value: 5}
	basicLimits.ClusterLimits.Egress = &UsageLimitValue{Unit: "MBPS", Value: 15}
	usageLimits.TierLimits["BASIC"] = basicLimits

	// Set Standard eCKU limits
	standardLimits := usageLimits.TierLimits["STANDARD"]
	standardLimits.ClusterLimits.MaxEcku = &UsageLimitValue{Value: 10}
	standardLimits.ClusterLimits.Ingress = &UsageLimitValue{Unit: "MBPS", Value: 25}
	standardLimits.ClusterLimits.Egress = &UsageLimitValue{Unit: "MBPS", Value: 75}
	usageLimits.TierLimits["STANDARD"] = standardLimits

	return usageLimits
}
