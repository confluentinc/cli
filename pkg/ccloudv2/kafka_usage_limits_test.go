package ccloudv2

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/config"
)

func TestUsageLimitsError(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "with error",
			input:    errors.New("test error"),
			expected: "usage limits API HTTP request failed: test error",
		},
		{
			name:     "with string",
			input:    "failed to get auth token",
			expected: "usage limits API HTTP request failed: failed to get auth token",
		},
		{
			name:     "with int",
			input:    404,
			expected: "usage limits API HTTP request failed: 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := usageLimitsError(tt.input)
			require.Error(t, err)
			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

func TestLimits_GetIngress(t *testing.T) {
	tests := []struct {
		name     string
		limits   *Limits
		expected int32
	}{
		{
			name:     "nil limits",
			limits:   nil,
			expected: 0,
		},
		{
			name:     "nil ingress",
			limits:   &Limits{Ingress: nil},
			expected: 0,
		},
		{
			name:     "valid ingress",
			limits:   &Limits{Ingress: &UsageLimitValue{Value: 100}},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetIngress()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLimits_GetEgress(t *testing.T) {
	tests := []struct {
		name     string
		limits   *Limits
		expected int32
	}{
		{
			name:     "nil limits",
			limits:   nil,
			expected: 0,
		},
		{
			name:     "nil egress",
			limits:   &Limits{Egress: nil},
			expected: 0,
		},
		{
			name:     "valid egress",
			limits:   &Limits{Egress: &UsageLimitValue{Value: 200}},
			expected: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetEgress()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLimits_GetStorage(t *testing.T) {
	storage := &UsageLimitValue{Value: 5000, Unit: "GB"}

	tests := []struct {
		name     string
		limits   *Limits
		expected *UsageLimitValue
	}{
		{
			name:     "nil limits",
			limits:   nil,
			expected: nil,
		},
		{
			name:     "nil storage",
			limits:   &Limits{Storage: nil},
			expected: nil,
		},
		{
			name:     "valid storage",
			limits:   &Limits{Storage: storage},
			expected: storage,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetStorage()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLimits_GetMaxEcku(t *testing.T) {
	maxEcku := &UsageLimitValue{Value: 10}

	tests := []struct {
		name     string
		limits   *Limits
		expected *UsageLimitValue
	}{
		{
			name:     "nil limits",
			limits:   nil,
			expected: nil,
		},
		{
			name:     "nil maxEcku",
			limits:   &Limits{MaxEcku: nil},
			expected: nil,
		},
		{
			name:     "valid maxEcku",
			limits:   &Limits{MaxEcku: maxEcku},
			expected: maxEcku,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetMaxEcku()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTierLimit_GetClusterLimits(t *testing.T) {
	clusterLimits := Limits{
		Ingress: &UsageLimitValue{Value: 100},
		Egress:  &UsageLimitValue{Value: 200},
	}

	tests := []struct {
		name      string
		tierLimit *TierLimit
		expected  *Limits
	}{
		{
			name:      "nil tierLimit",
			tierLimit: nil,
			expected:  nil,
		},
		{
			name:      "valid clusterLimits",
			tierLimit: &TierLimit{ClusterLimits: clusterLimits},
			expected:  &clusterLimits,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tierLimit.GetClusterLimits()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.Ingress.Value, result.GetIngress())
				assert.Equal(t, tt.expected.Egress.Value, result.GetEgress())
			}
		})
	}
}

func TestUsageLimits_GetCkuLimit(t *testing.T) {
	tests := []struct {
		name     string
		limits   *UsageLimits
		expected *Limits
	}{
		{
			name:     "nil usageLimits",
			limits:   nil,
			expected: nil,
		},
		{
			name: "cku not found",
			limits: &UsageLimits{
				CkuLimits: map[string]Limits{},
			},
			expected: nil,
		},
		{
			name: "valid cku",
			limits: &UsageLimits{
				CkuLimits: map[string]Limits{"1": {Ingress: &UsageLimitValue{Value: 60}}}},
			expected: &Limits{Ingress: &UsageLimitValue{Value: 60}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetCkuLimit(1)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected.GetIngress(), result.GetIngress())
			}
		})
	}
}

func TestUsageLimits_GetTierLimit(t *testing.T) {
	tests := []struct {
		name     string
		limits   *UsageLimits
		sku      string
		expected *TierLimit
	}{
		{
			name: "nil usageLimits",
			sku:  "BASIC",
		},
		{
			name:   "sku not found",
			limits: &UsageLimits{TierLimits: map[string]TierLimit{}},
			sku:    "BASIC",
		},
		{
			name: "valid tier",
			limits: &UsageLimits{
				TierLimits: map[string]TierLimit{
					"BASIC": {
						ClusterLimits: Limits{
							Ingress: &UsageLimitValue{Value: 25},
						},
					},
				},
			},
			sku: "BASIC",
			expected: &TierLimit{
				ClusterLimits: Limits{
					Ingress: &UsageLimitValue{Value: 25},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetTierLimit(tt.sku)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				clusterLimits := result.GetClusterLimits()
				require.NotNil(t, clusterLimits)
				assert.Equal(t, tt.expected.ClusterLimits.GetIngress(), clusterLimits.GetIngress())
			}
		})
	}
}

func TestUsageLimits_GetUsageLimits(t *testing.T) {
	validResponse := UsageLimitsResponse{
		UsageLimits: UsageLimits{
			TierLimits: map[string]TierLimit{
				"BASIC": {ClusterLimits: Limits{Ingress: &UsageLimitValue{Value: 25, Unit: "MBPS"}}},
			},
		},
	}
	testToken := "test-token"

	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter)
		authToken      string
		expectedError  string
		expectedResult *UsageLimits
	}{
		{
			name: "successful GetUsageLimits request",
			serverResponse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusOK)
				err := json.NewEncoder(w).Encode(validResponse)
				require.NoError(t, err)
			},
			authToken:      testToken,
			expectedResult: &validResponse.UsageLimits,
		},
		{
			name:          "GetUsageLimits request with missing auth token",
			authToken:     "",
			expectedError: ErrFailedToGetAuthToken,
		},
		{
			name: "GetUsageLimits request with invalid JSON response",
			serverResponse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("invalid json"))
				require.NoError(t, err)
			},
			authToken:     testToken,
			expectedError: ErrFailedToDecodeResponse,
		},
		{
			name: "GetUsageLimits request with API error response",
			serverResponse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusOK)
				errorMsg := "API error message"
				err := json.NewEncoder(w).Encode(UsageLimitsResponse{Error: &errorMsg})
				require.NoError(t, err)
			},
			authToken:     testToken,
			expectedError: "API error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			serverURL := "https://api.confluent.cloud"
			if tt.serverResponse != nil {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "GET", r.Method)
					assert.Equal(t, "Bearer "+tt.authToken, r.Header.Get("Authorization"))
					tt.serverResponse(w)
				}))
				defer server.Close()
				serverURL = server.URL
			}

			client := &Client{cfg: &config.Config{
				CurrentContext: "context",
				Contexts: map[string]*config.Context{"context": {
					Name:           "context",
					Platform:       &config.Platform{Name: "test-platform", Server: serverURL},
					Credential:     &config.Credential{Name: "test-credential"},
					State:          &config.ContextState{AuthToken: tt.authToken},
					PlatformName:   "test-platform",
					CredentialName: "test-credential",
				}},
			}}

			result, err := client.GetUsageLimits("aws", "lkc-123", "env-456")

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.NotNil(t, result)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}
