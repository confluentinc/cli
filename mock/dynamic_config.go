package mock

import (
	"github.com/confluentinc/cli/v3/pkg/config"
	dynamicconfig "github.com/confluentinc/cli/v3/pkg/dynamic-config"
	"github.com/confluentinc/cli/v3/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *dynamicconfig.DynamicConfig {
	cfg := config.AuthenticatedCloudConfigMock()
	v2Client := mock.NewV2ClientMock()
	return dynamicconfig.New(cfg, v2Client)
}
