package schema_registry

import (
	"github.com/confluentinc/ccloudapis/org/v1"
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"testing"
)

func TestSrContextFound(t *testing.T) {
	ctx, err := srContext(&config.Config{
		CurrentContext: "ctx",
		Auth: &config.AuthConfig{Account: &v1.Account{Id: "me"}},
		Contexts: map[string]*config.Context{"ctx": {
			SchemaRegistryClusters: map[string]*config.SchemaRegistryCluster{
				"me": {
					SrCredentials: &config.APIKeyPair{
						Key:    "aladdin",
						Secret: "opensesame",
					},
				},
			},
		}},
	})
	if err != nil || ctx.Value(srsdk.ContextBasicAuth) == nil {
		t.Fail()
	}
}
