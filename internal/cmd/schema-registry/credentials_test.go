package schema_registry

import (
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/config"
)

func TestSrContextFound(t *testing.T) {
	ctx, err := srContext(config.AuthenticatedConfigMock())
	if err != nil || ctx.Value(srsdk.ContextBasicAuth) == nil {
		t.Fail()
	}
}
