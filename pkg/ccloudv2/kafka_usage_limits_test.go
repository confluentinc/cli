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

var (
	testmMaxEcku      = &UsageLimitValue{Value: 10}
	testStorage       = &UsageLimitValue{Value: 5000, Unit: "GB"}
	testIngress       = &UsageLimitValue{Value: 100, Unit: "MBPS"}
	testEgress        = &UsageLimitValue{Value: 150, Unit: "MBPS"}
	testClusterLimits = Limits{Ingress: testIngress, Egress: testEgress, Storage: testStorage, MaxEcku: testmMaxEcku}
	testTierLimits    = map[string]TierLimit{"BASIC": {ClusterLimits: testClusterLimits}}
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

func TestLimitsGetters(t *testing.T) {
	tests := []struct {
		name            string
		limits          *Limits
		expectedIngress int32
		expectedEgress  int32
		expectedStorage *UsageLimitValue
		expectedMaxEcku *UsageLimitValue
	}{
		{
			name:            "nil clusters limits",
			limits:          nil,
			expectedIngress: 0,
			expectedEgress:  0,
			expectedStorage: nil,
			expectedMaxEcku: nil,
		},
		{
			name:            "nil limits values",
			limits:          &Limits{},
			expectedIngress: 0,
			expectedEgress:  0,
			expectedStorage: nil,
			expectedMaxEcku: nil,
		},
		{
			name:            "valid limits values",
			limits:          &testClusterLimits,
			expectedIngress: 100,
			expectedEgress:  150,
			expectedStorage: testStorage,
			expectedMaxEcku: testmMaxEcku,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedIngress, tt.limits.GetIngress())
			assert.Equal(t, tt.expectedEgress, tt.limits.GetEgress())
			assert.Equal(t, tt.expectedStorage, tt.limits.GetStorage())
			assert.Equal(t, tt.expectedMaxEcku, tt.limits.GetMaxEcku())
		})
	}
}

func TestTierLimitGetClusterLimits(t *testing.T) {
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
			tierLimit: &TierLimit{ClusterLimits: testClusterLimits},
			expected:  &testClusterLimits,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tierLimit.GetClusterLimits()
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUsageLimitsGetCkuLimit(t *testing.T) {
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
			name:     "cku not found",
			limits:   &UsageLimits{CkuLimits: map[string]Limits{}},
			expected: nil,
		},
		{
			name:     "valid cku",
			limits:   &UsageLimits{CkuLimits: map[string]Limits{"1": {Ingress: testIngress}}},
			expected: &Limits{Ingress: testIngress},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.limits.GetCkuLimit(1)
			if tt.expected == nil {
				assert.Nil(t, result)
			} else {
				require.NotNil(t, result)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestUsageLimitsGetTierLimit(t *testing.T) {
	tests := []struct {
		name     string
		limits   *UsageLimits
		sku      string
		expected *TierLimit
	}{
		{
			name:     "nil usageLimits",
			sku:      "BASIC",
			expected: nil,
		},
		{
			name:     "sku not found",
			limits:   &UsageLimits{TierLimits: map[string]TierLimit{}},
			sku:      "BASIC",
			expected: nil,
		},
		{
			name:     "valid tier limits",
			limits:   &UsageLimits{TierLimits: testTierLimits},
			sku:      "BASIC",
			expected: &TierLimit{testClusterLimits},
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
				assert.Equal(t, tt.expected.ClusterLimits, *clusterLimits)
			}
		})
	}
}

func TestUsageLimitsGetUsageLimits(t *testing.T) {
	validResponse := UsageLimitsResponse{UsageLimits: UsageLimits{TierLimits: testTierLimits}}
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
			expectedError: FailedToGetAuthTokenErrorMsg,
		},
		{
			name: "GetUsageLimits request with invalid JSON response",
			serverResponse: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusOK)
				_, err := w.Write([]byte("invalid json"))
				require.NoError(t, err)
			},
			authToken:     testToken,
			expectedError: FailedToDecodeResponseErrorMsg,
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
