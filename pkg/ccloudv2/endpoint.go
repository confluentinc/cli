package ccloudv2

import (
	"context"
	"net/http"
	"strings"

	endpointv1 "github.com/confluentinc/ccloud-sdk-go-v2/endpoint/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newEndpointClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *endpointv1.APIClient {
	cfg := endpointv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = endpointv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return endpointv1.NewAPIClient(cfg)
}

func (c *Client) endpointApiContext() context.Context {
	return context.WithValue(context.Background(), endpointv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// ListEndpoints returns all endpoints matching the given filters, paginating as
// needed. The `cloud` and `service` filters are normalized to uppercase to match
// the API's case-sensitive filter semantics (e.g. AWS, FLINK), so callers can
// pass either case without surprise.
func (c *Client) ListEndpoints(environment, cloud, region, service string, isPrivate *bool, resource string) ([]endpointv1.EndpointV1Endpoint, error) {
	cloud = strings.ToUpper(cloud)
	service = strings.ToUpper(service)

	var list []endpointv1.EndpointV1Endpoint

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListEndpoints(environment, pageToken, cloud, region, service, isPrivate, resource)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		pageToken, done, err = extractNextPageToken(page.GetMetadata().Next)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

func (c *Client) executeListEndpoints(environment, pageToken, cloud, region, service string, isPrivate *bool, resource string) (endpointv1.EndpointV1EndpointList, *http.Response, error) {
	req := c.EndpointClient.EndpointsEndpointV1Api.ListEndpointV1Endpoints(c.endpointApiContext()).
		Environment(environment).
		PageSize(ccloudV2ListPageSize)

	if service != "" {
		req = req.Service(service)
	}
	if cloud != "" {
		req = req.Cloud(cloud)
	}
	if region != "" {
		req = req.Region(region)
	}
	if isPrivate != nil {
		req = req.IsPrivate(*isPrivate)
	}
	if resource != "" {
		req = req.Resource(resource)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}

	return req.Execute()
}
