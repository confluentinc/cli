//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/ccloud_client_factory.go --pkg mock --selfpkg github.com/confluentinc/cli ccloud_client_factory.go CCloudClientFactory
package auth

import (
	"context"

	"github.com/confluentinc/ccloud-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type CCloudClientFactory interface {
	AnonHTTPClientFactory(baseURL string) *ccloud.Client
	JwtHTTPClientFactory(ctx context.Context, jwt string, baseURL string) *ccloud.Client
}

type CCloudClientFactoryImpl struct {
	UserAgent string
	Logger    *log.Logger
}

func NewCCloudClientFactory(userAgent string, logger *log.Logger) CCloudClientFactory {
	return &CCloudClientFactoryImpl{
		UserAgent: userAgent,
		Logger:    logger,
	}
}

func (c *CCloudClientFactoryImpl) AnonHTTPClientFactory(baseURL string) *ccloud.Client {
	return ccloud.NewClient(&ccloud.Params{BaseURL: baseURL, HttpClient: ccloud.BaseClient, Logger: c.Logger, UserAgent: c.UserAgent})
}

func (c *CCloudClientFactoryImpl) JwtHTTPClientFactory(ctx context.Context, jwt string, baseURL string) *ccloud.Client {
	return ccloud.NewClientWithJWT(ctx, jwt, &ccloud.Params{BaseURL: baseURL, Logger: c.Logger, UserAgent: c.UserAgent})
}
