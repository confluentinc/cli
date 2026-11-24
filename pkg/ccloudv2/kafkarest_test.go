package ccloudv2

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDeleteKafkaConsumerGroup(t *testing.T) {
	tests := []struct {
		name         string
		groupId      string
		status       int
		body         string
		wantRawPath  string
		wantErr      string
		wantSuggests bool
	}{
		{
			name:        "empty group is deleted",
			groupId:     "my-group",
			status:      http.StatusNoContent,
			wantRawPath: "/kafka/v3/clusters/lkc-123/internal/consumer-groups/my-group",
		},
		{
			name:        "group id is path-escaped",
			groupId:     "team/a b",
			status:      http.StatusNoContent,
			wantRawPath: "/kafka/v3/clusters/lkc-123/internal/consumer-groups/team%2Fa%20b",
		},
		{
			name:        "group with active members",
			groupId:     "my-group",
			status:      http.StatusBadRequest,
			body:        `{"error_code":400,"message":"my-group has state STABLE, delete only allowed in state 'empty' or 'dead'."}`,
			wantRawPath: "/kafka/v3/clusters/lkc-123/internal/consumer-groups/my-group",
			wantErr:     "REST request failed: my-group has state STABLE, delete only allowed in state 'empty' or 'dead'.",
		},
		{
			name:        "unknown group",
			groupId:     "my-group",
			status:      http.StatusNotFound,
			body:        `{"error_code":40405,"message":"This server does not host this consumer group."}`,
			wantRawPath: "/kafka/v3/clusters/lkc-123/internal/consumer-groups/my-group",
			wantErr:     "REST request failed: This server does not host this consumer group.",
		},
		{
			name:         "error without a JSON body",
			groupId:      "my-group",
			status:       http.StatusBadRequest,
			body:         "<html>bad request</html>",
			wantRawPath:  "/kafka/v3/clusters/lkc-123/internal/consumer-groups/my-group",
			wantErr:      "Kafka REST request failed",
			wantSuggests: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				require.Equal(t, http.MethodDelete, r.Method)
				require.Equal(t, test.wantRawPath, r.URL.EscapedPath())
				require.Equal(t, "Bearer token-abc", r.Header.Get("Authorization"))
				require.Equal(t, "application/json", r.Header.Get("Accept"))
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			client := NewKafkaRestClient(server.URL+"/", "lkc-123", "test-agent", "token-abc", false)
			err := client.DeleteKafkaConsumerGroup(test.groupId)

			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestKafkaConsumerGroupExists(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		body       string
		wantExists bool
		wantErr    bool
	}{
		{name: "group exists", status: http.StatusOK, body: `{"kind":"KafkaConsumerGroup","consumer_group_id":"my-group"}`, wantExists: true},
		{name: "group not found", status: http.StatusNotFound, body: `{"error_code":40403,"message":"This server does not host this consumer group."}`},
		{name: "describe forbidden", status: http.StatusForbidden, body: `{"error_code":40301,"message":"Not authorized."}`, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(test.status)
				_, _ = w.Write([]byte(test.body))
			}))
			defer server.Close()

			client := NewKafkaRestClient(server.URL+"/", "lkc-123", "test-agent", "token-abc", false)
			exists, err := client.KafkaConsumerGroupExists("my-group")

			require.Equal(t, test.wantExists, exists)
			if test.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
