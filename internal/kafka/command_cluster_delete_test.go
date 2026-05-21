package kafka

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDeletionProtectionErrDetail(t *testing.T) {
	tests := []struct {
		name           string
		response       *http.Response
		expectedDetail string
		expectedOk     bool
	}{
		{
			name:       "nil response",
			response:   nil,
			expectedOk: false,
		},
		{
			name: "non-conflict status code",
			response: &http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "deletion_protection_enabled", "detail": "test"}]}`)),
			},
			expectedOk: false,
		},
		{
			name: "conflict with deletion protection error code",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "deletion_protection_enabled", "detail": "Cluster deletion is blocked by deletion protection."}]}`)),
			},
			expectedDetail: "Cluster deletion is blocked by deletion protection.",
			expectedOk:     true,
		},
		{
			name: "conflict with different error code",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "some_other_error", "detail": "some detail"}]}`)),
			},
			expectedOk: false,
		},
		{
			name: "conflict with empty errors array",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader(`{"errors": []}`)),
			},
			expectedOk: false,
		},
		{
			name: "conflict with invalid JSON body",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader(`not json`)),
			},
			expectedOk: false,
		},
		{
			name: "conflict with nil body",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       nil,
			},
			expectedOk: false,
		},
		{
			name: "deletion protection error is not first in errors array",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body: io.NopCloser(strings.NewReader(`{"errors": [` +
					`{"code": "some_other_error", "detail": "other error"},` +
					`{"code": "deletion_protection_enabled", "detail": "Cluster deletion is blocked by deletion protection."}` +
					`]}`)),
			},
			expectedDetail: "Cluster deletion is blocked by deletion protection.",
			expectedOk:     true,
		},
		{
			name: "body is restored after reading",
			response: &http.Response{
				StatusCode: http.StatusConflict,
				Body:       io.NopCloser(strings.NewReader(`{"errors": [{"code": "deletion_protection_enabled", "detail": "Cluster deletion is blocked by deletion protection."}]}`)),
			},
			expectedDetail: "Cluster deletion is blocked by deletion protection.",
			expectedOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			detail, ok := parseDeletionProtectionErrDetail(tt.response)
			require.Equal(t, tt.expectedOk, ok)
			require.Equal(t, tt.expectedDetail, detail)

			// Verify body is restored for downstream handlers
			if tt.response != nil && tt.response.Body != nil && ok {
				body, err := io.ReadAll(tt.response.Body)
				require.NoError(t, err)
				require.NotEmpty(t, body)
			}
		})
	}
}

func TestDeletionProtectionErrorToSuggestion(t *testing.T) {
	tests := []struct {
		name               string
		errorMsg           string
		expectedSuggestion string
	}{
		{
			name:               "cluster deletion protection",
			errorMsg:           "Cluster deletion is blocked by deletion protection.",
			expectedSuggestion: `Disable deletion_protection before deleting the cluster.`,
		},
		{
			name:               "cluster deletion protection case insensitive",
			errorMsg:           "cluster deletion is blocked by deletion protection.",
			expectedSuggestion: `Disable deletion_protection before deleting the cluster.`,
		},
		{
			name:               "unknown deletion protection error",
			errorMsg:           "Some other deletion protection error.",
			expectedSuggestion: "",
		},
		{
			name:               "empty string",
			errorMsg:           "",
			expectedSuggestion: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestion := deletionProtectionErrorToSuggestion(tt.errorMsg)
			require.Equal(t, tt.expectedSuggestion, suggestion)
		})
	}
}
