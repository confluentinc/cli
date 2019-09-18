package schema_registry

import (
	"testing"

	"github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
)

func TestSrHasApiKey(t *testing.T) {
	hasApiKey := srHasAPIKey(&config.Config{
		CurrentContext: "ctx",
		Auth:           &config.AuthConfig{Account: &v1.Account{Id: "me"}},
		Contexts: map[string]*config.Context{"ctx": {
			SchemaRegistryClusters: map[string]*config.SchemaRegistryCluster{
				"me": {
					SrCredentials: &config.APIKeyPair{
						Key:    "",
						Secret: "",
					},
				},
			},
		}},
	})
	if hasApiKey {
		t.Fail()
	}
}
