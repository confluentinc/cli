package schema_registry

import (
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/mock"
)

func TestSrContextFound(t *testing.T) {
	cfg := mock.AuthenticatedDynamicConfigMock()
	cmd := &cobra.Command{}
	ctx, err := srContext(cfg, cmd)
	if err != nil || ctx.Value(srsdk.ContextBasicAuth) == nil {
		t.Fail()
	}
}
