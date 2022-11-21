package mock

import (
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/mock"
)

func AuthenticatedDynamicConfigMock() *dynamicconfig.DynamicConfig {
	cfg := v1.AuthenticatedCloudConfigMock()
	privateClient := mock.NewPrivateClientMock()
	client := mock.NewClientMock()
	v2Client := mock.NewV2ClientMock()
	return dynamicconfig.New(cfg, privateClient, client, v2Client)
}
