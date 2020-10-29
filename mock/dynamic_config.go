package mock

import (
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *pcmd.DynamicConfig {
	cfg := v3.AuthenticatedCloudConfigMock()
	client := mock.NewClientMock()
	flagResolverMock := &pcmd.FlagResolverImpl{
		Prompt: &mock.Prompt{},
		Out:    os.Stdout,
	}
	return pcmd.NewDynamicConfig(cfg, flagResolverMock, client)
}
