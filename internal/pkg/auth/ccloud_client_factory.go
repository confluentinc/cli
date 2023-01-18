//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/ccloud_client_factory.go --pkg mock --selfpkg github.com/confluentinc/cli ccloud_client_factory.go CCloudClientFactory
package auth

import (
	"context"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/log"
)

type CCloudClientFactory interface {
	AnonHTTPClientFactory(baseURL string) *ccloudv1.Client
	JwtHTTPClientFactory(ctx context.Context, jwt string, baseURL string) *ccloudv1.Client
}

type CCloudClientFactoryImpl struct {
	UserAgent string
}

func NewCCloudClientFactory(userAgent string) CCloudClientFactory {
	return &CCloudClientFactoryImpl{
		UserAgent: userAgent,
	}
}

func (c *CCloudClientFactoryImpl) AnonHTTPClientFactory(baseURL string) *ccloudv1.Client {
	return ccloudv1.NewClient(&ccloudv1.Params{BaseURL: baseURL, HttpClient: ccloudv1.BaseClient, Logger: log.CliLogger, UserAgent: c.UserAgent})
}

func (c *CCloudClientFactoryImpl) JwtHTTPClientFactory(ctx context.Context, jwt string, baseURL string) *ccloudv1.Client {
	return ccloudv1.NewClientWithJWT(ctx, jwt, &ccloudv1.Params{BaseURL: baseURL, Logger: log.CliLogger, UserAgent: c.UserAgent})
}
