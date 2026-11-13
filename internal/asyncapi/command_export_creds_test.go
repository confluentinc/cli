package asyncapi

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	apikeysmock "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2/mock"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newCredsTestCommand(ctx *config.Context, getApiKey func(apikeysv2.ApiGetIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error)) *command {
	api := &apikeysmock.APIKeysIamV2Api{}
	api.GetIamV2ApiKeyFunc = func(_ context.Context, _ string) apikeysv2.ApiGetIamV2ApiKeyRequest {
		return apikeysv2.ApiGetIamV2ApiKeyRequest{ApiService: api}
	}
	api.GetIamV2ApiKeyExecuteFunc = getApiKey
	// GetApiKey falls back to listing keys when the lookup fails
	api.ListIamV2ApiKeysFunc = func(_ context.Context) apikeysv2.ApiListIamV2ApiKeysRequest {
		return apikeysv2.ApiListIamV2ApiKeysRequest{ApiService: api}
	}
	api.ListIamV2ApiKeysExecuteFunc = func(_ apikeysv2.ApiListIamV2ApiKeysRequest) (apikeysv2.IamV2ApiKeyList, *http.Response, error) {
		return apikeysv2.IamV2ApiKeyList{}, nil, nil
	}

	return &command{AuthenticatedCLICommand: &pcmd.AuthenticatedCLICommand{
		Context:  ctx,
		V2Client: &ccloudv2.Client{ApiKeysClient: &apikeysv2.APIClient{APIKeysIamV2Api: api}},
	}}
}

func apiKeyForResource(id string) func(apikeysv2.ApiGetIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error) {
	return func(_ apikeysv2.ApiGetIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error) {
		return apikeysv2.IamV2ApiKey{Spec: &apikeysv2.IamV2ApiKeySpec{
			Resource: *apikeysv2.NewNullableTypedEnvScopedObjectReference(&apikeysv2.TypedEnvScopedObjectReference{Id: id}),
		}}, nil, nil
	}
}

func TestGetClusterCreds_ExplicitKeyStoredLocally(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, nil)
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456", APIKeys: map[string]*config.APIKeyPair{"KEY": {Key: "KEY", Secret: "SECRET"}}}

	creds, err := c.getClusterCreds(cluster, "KEY")
	require.NoError(t, err)
	require.Equal(t, &config.APIKeyPair{Key: "KEY", Secret: "SECRET"}, creds)
}

func TestGetClusterCreds_ExplicitKeyLookupFails(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, func(_ apikeysv2.ApiGetIamV2ApiKeyRequest) (apikeysv2.IamV2ApiKey, *http.Response, error) {
		return apikeysv2.IamV2ApiKey{}, nil, fmt.Errorf("not found")
	})
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456"}

	_, err := c.getClusterCreds(cluster, "KEY")
	require.EqualError(t, err, "unknown API key KEY")
}

func TestGetClusterCreds_ExplicitKeyForOtherCluster(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, apiKeyForResource("lkc-other"))
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456"}

	_, err := c.getClusterCreds(cluster, "KEY")
	require.EqualError(t, err, fmt.Sprintf(errors.InvalidApiKeyErrorMsg, "KEY", "lkc-123456"))
}

func TestGetClusterCreds_ExplicitKeySecretNotStored(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, apiKeyForResource("lkc-123456"))
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456"}

	_, err := c.getClusterCreds(cluster, "KEY")
	require.Equal(t, &errors.UnconfiguredAPISecretError{APIKey: "KEY", ClusterID: "lkc-123456"}, err)
}

func TestGetClusterCreds_ActiveClusterKey(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, nil)
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456", APIKey: "KEY", APIKeys: map[string]*config.APIKeyPair{"KEY": {Key: "KEY", Secret: "SECRET"}}}

	creds, err := c.getClusterCreds(cluster, "")
	require.NoError(t, err)
	require.Equal(t, &config.APIKeyPair{Key: "KEY", Secret: "SECRET"}, creds)
}

func TestGetClusterCreds_ActiveClusterKeySecretMissing(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, nil)
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456", APIKey: "KEY"}

	_, err := c.getClusterCreds(cluster, "")
	require.Error(t, err)
}

func TestGetClusterCreds_NoKeySet(t *testing.T) {
	c := newCredsTestCommand(&config.Context{}, nil)
	cluster := &config.KafkaClusterConfig{ID: "lkc-123456"}

	_, err := c.getClusterCreds(cluster, "")
	require.EqualError(t, err, "API key not set for the Kafka cluster")
}
