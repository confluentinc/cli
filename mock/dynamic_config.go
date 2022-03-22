package mock

import (
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *pcmd.DynamicConfig {
	cfg := v1.AuthenticatedCloudConfigMock()
	client := mock.NewClientMock()
	v2Client := mock.NewV2ClientMock()
	flagResolverMock := &pcmd.FlagResolverImpl{
		Prompt: &mock.Prompt{},
		Out:    os.Stdout,
	}
	return pcmd.NewDynamicConfig(cfg, flagResolverMock, client, v2Client)
}
