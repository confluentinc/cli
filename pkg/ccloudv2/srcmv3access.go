package ccloudv2

import (
	"context"
	"net/http"

	srcmv3access "github.com/confluentinc/ccloud-sdk-go-v2/srcmv3access/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newSrcmV3AccessClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *srcmv3access.APIClient {
	cfg := srcmv3access.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = srcmv3access.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return srcmv3access.NewAPIClient(cfg)
}

func (c *Client) srcmV3AccessApiContext() context.Context {
	return context.WithValue(context.Background(), srcmv3access.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) GetSchemaRegistryV3AccessById(clusterId, environment string) (srcmv3access.SrcmV3Access, error) {
	cluster, httpResp, err := c.SrcmV3AcessClient.AccessesSrcmV3Api.GetSrcmV3Access(c.srcmV3AccessApiContext(), clusterId).Environment(environment).ClusterId(clusterId).Execute()
	return cluster, errors.CatchCCloudV2Error(err, httpResp)
}
