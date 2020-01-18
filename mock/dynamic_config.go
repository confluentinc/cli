package mock

import (
	"os"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *cmd.DynamicConfig {
	cfg := config.AuthenticatedConfigMock()
	client := mock.NewClientMock()
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return cmd.NewDynamicConfig(cfg, flagResolverMock, client)
}