//go:generate go run go.uber.org/mock/mockgen -typed -destination=../../mock/ccloud_client_factory.go -package=mock github.com/confluentinc/cli/v4/pkg/auth CCloudClientFactory
package auth

import (
	"context"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v4/pkg/httpclient"
	"github.com/confluentinc/cli/v4/pkg/log"
)

type CCloudClientFactory interface {
	AnonHTTPClientFactory(baseURL string) *ccloudv1.Client
	JwtHTTPClientFactory(ctx context.Context, jwt, baseURL string) *ccloudv1.Client
}

type CCloudClientFactoryImpl struct {
	UserAgent string
}

func NewCCloudClientFactory(userAgent string) CCloudClientFactory {
	return &CCloudClientFactoryImpl{UserAgent: userAgent}
}

func (c *CCloudClientFactoryImpl) AnonHTTPClientFactory(baseURL string) *ccloudv1.Client {
	return ccloudv1.NewClient(&ccloudv1.Params{BaseURL: baseURL, HttpClient: httpclient.New(httpclient.Default), Logger: log.CliLogger, UserAgent: c.UserAgent})
}

func (c *CCloudClientFactoryImpl) JwtHTTPClientFactory(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
	return ccloudv1.NewClientWithJWT(ctx, jwt, &ccloudv1.Params{BaseURL: baseURL, BaseClient: httpclient.New(httpclient.Default), Logger: log.CliLogger, UserAgent: c.UserAgent})
}
