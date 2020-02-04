package mock

import (
	"os"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *cmd.DynamicConfig {
	cfg := v1.AuthenticatedConfigMock()
	client := mock.NewClientMock()
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return cmd.NewDynamicConfig(cfg, flagResolverMock, client)
}