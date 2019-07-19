package schema_registry

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"testing"
)

func TestSrContextFound(t *testing.T) {
	ctx, err := srContext(&config.Config{
		SrCredentials: &config.APIKeyPair{
			Key:    "aladdin",
			Secret: "opensesame",
		},
	})
	if err != nil || ctx.Value(srsdk.ContextBasicAuth) == nil {
		t.Fail()
	}
}
