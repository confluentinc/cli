package test_server

import (
	"encoding/json"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"testing"
)

func handleKafkaACLsList(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		results := []*schedv1.ACLBinding{
			{
				Pattern: &schedv1.ResourcePatternConfig{
					ResourceType: schedv1.ResourceTypes_TOPIC,
					Name:         "test-topic",
					PatternType:  schedv1.PatternTypes_LITERAL,
				},
				Entry: &schedv1.AccessControlEntryConfig{
					Operation:      schedv1.ACLOperations_READ,
					PermissionType: schedv1.ACLPermissionTypes_ALLOW,
				},
			},
		}
		reply, err := json.Marshal(results)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}
