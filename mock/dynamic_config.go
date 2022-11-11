package mock

import (
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *dynamicconfig.DynamicConfig {
	cfg := v1.AuthenticatedCloudConfigMock()
	client := mock.NewClientMock()
	publicClient := mock.NewPublicClientMock()
	v2Client := mock.NewV2ClientMock()
	return dynamicconfig.New(cfg, client, publicClient, v2Client)
}
