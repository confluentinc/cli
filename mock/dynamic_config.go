package mock

import (
	"github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *dynamic_config.DynamicConfig {
	cfg := v1.AuthenticatedCloudConfigMock()
	client := mock.NewClientMock()
	v2Client := mock.NewV2ClientMock()
	flagResolverMock := &pcmd.FlagResolverImpl{
		Prompt: &mock.Prompt{},
		Out:    os.Stdout,
	}
	return dynamic_config.NewDynamicConfig(cfg, flagResolverMock, client, v2Client)
}
