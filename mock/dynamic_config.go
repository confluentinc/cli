package mock

import (
	"os"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *cmd.DynamicConfig {
	cfg := v2.AuthenticatedConfigMock()
	client := mock.NewClientMock()
	flagResolverMock := &cmd.FlagResolverImpl{
		Prompt: &Prompt{},
		Out:    os.Stdout,
	}
	return cmd.NewDynamicConfig(cfg, flagResolverMock, client)
}
