package ccloudv2

import (
	"context"
	"net/http"

	switchoverv1 "github.com/confluentinc/ccloud-sdk-go-v2/switchover/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func newSwitchoverClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *switchoverv1.APIClient {
	cfg := switchoverv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = switchoverv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return switchoverv1.NewAPIClient(cfg)
}

// switchoverApiContext selects the credential the Switchover API is called with.
//
// The Switchover API is served by frontdoor-api-gateway, which accepts Cloud /
// Global API keys and regional customer access tokens but not raw login-session
// JWTs. Credentials are resolved from the CLI's existing keystore and login
// state, in this order:
//
//  1. The active Global API key in the local keystore (set by
//     'api-key create --resource global' or 'api-key use <key>'), sent as HTTP
//     Basic auth. A key the user explicitly selected always wins over ambient
//     login state, so the acting principal can be pinned (for example to a
//     service account for a failover) and the token exchange bypassed without
//     logging out. This is also the break-glass path when the exchange is down.
//  2. The login session: the session token is exchanged for a regional customer
//     access token via auth.GetRegionalToken and sent as a bearer. This is the
//     default for 'confluent login' users, matching how the other CLI clients
//     authenticate to token-exchange-backed APIs.
//  3. A Cloud API key stored in the context by an API-key login, as Basic auth.
//  4. The raw session token. frontdoor rejects it today, but it keeps the error
//     surfaced as a 401 from the API rather than a silent unauthenticated request,
//     and it will start working if frontdoor ever accepts session JWTs.
func (c *Client) switchoverApiContext() context.Context {
	ctx := c.cfg.Context()
	if pair := ctx.GetActiveGlobalAPIKeyPair(); pair != nil {
		if err := pair.DecryptSecret(); err == nil {
			return context.WithValue(context.Background(), switchoverv1.ContextBasicAuth, switchoverv1.BasicAuth{UserName: pair.Key, Password: pair.Secret})
		} else {
			log.CliLogger.Debugf("switchover: could not decrypt active Global API key %q, falling back to login: %v", pair.Key, err)
		}
	}
	if ctx != nil && ctx.GetAuthToken() != "" {
		token, err := auth.GetRegionalToken(ctx)
		if err == nil {
			return context.WithValue(context.Background(), switchoverv1.ContextAccessToken, token)
		}
		log.CliLogger.Debugf("switchover: regional token exchange failed, falling back to stored API key: %v", err)
	}
	if ctx != nil && ctx.GetCredentialType() == config.APIKey && ctx.Credential.APIKeyPair != nil {
		pair := ctx.Credential.APIKeyPair
		if err := pair.DecryptSecret(); err == nil {
			return context.WithValue(context.Background(), switchoverv1.ContextBasicAuth, switchoverv1.BasicAuth{UserName: pair.Key, Password: pair.Secret})
		}
	}
	return context.WithValue(context.Background(), switchoverv1.ContextAccessToken, ctx.GetAuthToken())
}

func (c *Client) CreateSwitchoverPair(pair switchoverv1.SwitchoverV1SwitchoverPair) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.CreateSwitchoverV1SwitchoverPair(c.switchoverApiContext()).SwitchoverV1SwitchoverPair(pair).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSwitchoverPair(id, environment string) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.GetSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverPair(id, environment string, update switchoverv1.SwitchoverV1SwitchoverPairUpdateRequest) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).SwitchoverV1SwitchoverPairUpdateRequest(update).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) TriggerSwitchoverPairFailover(id string, req switchoverv1.SwitchoverV1SwitchoverPairFailoverRequest) (switchoverv1.SwitchoverV1SwitchoverPair, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.FailoverSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).SwitchoverV1SwitchoverPairFailoverRequest(req).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSwitchoverPairs(environment string) ([]switchoverv1.SwitchoverV1SwitchoverPair, error) {
	var list []switchoverv1.SwitchoverV1SwitchoverPair

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSwitchoverPairs(environment, pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		metadata := page.GetMetadata()
		pagination := metadata.GetPagination()
		pageToken = pagination.GetNextPageToken()
		done = pageToken == ""
	}

	return list, nil
}

func (c *Client) executeListSwitchoverPairs(environment, pageToken string) (switchoverv1.SwitchoverV1SwitchoverPairList, *http.Response, error) {
	req := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.ListSwitchoverV1SwitchoverPairs(c.switchoverApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteSwitchoverPair(id, environment string) error {
	_, httpResp, err := c.SwitchoverClient.SwitchoverPairsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverPair(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateSwitchoverEndpoint(endpoint switchoverv1.SwitchoverV1SwitchoverEndpoint) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.CreateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext()).SwitchoverV1SwitchoverEndpoint(endpoint).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetSwitchoverEndpoint(id, environment string) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.GetSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateSwitchoverEndpoint(id, environment string, update switchoverv1.SwitchoverV1SwitchoverEndpointUpdateRequest) (switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	res, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.UpdateSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).SwitchoverV1SwitchoverEndpointUpdateRequest(update).Execute()
	return res, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListSwitchoverEndpoints(environment, switchoverPair string) ([]switchoverv1.SwitchoverV1SwitchoverEndpoint, error) {
	var list []switchoverv1.SwitchoverV1SwitchoverEndpoint

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListSwitchoverEndpoints(environment, switchoverPair, pageToken)
		if err != nil {
			return nil, errors.CatchCCloudV2Error(err, httpResp)
		}
		list = append(list, page.GetData()...)

		metadata := page.GetMetadata()
		pagination := metadata.GetPagination()
		pageToken = pagination.GetNextPageToken()
		done = pageToken == ""
	}

	return list, nil
}

func (c *Client) executeListSwitchoverEndpoints(environment, switchoverPair, pageToken string) (switchoverv1.SwitchoverV1SwitchoverEndpointList, *http.Response, error) {
	req := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.ListSwitchoverV1SwitchoverEndpoints(c.switchoverApiContext()).Environment(environment).PageSize(ccloudV2ListPageSize)
	if switchoverPair != "" {
		req = req.SwitchoverPair(switchoverPair)
	}
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) DeleteSwitchoverEndpoint(id, environment string) error {
	_, httpResp, err := c.SwitchoverClient.SwitchoverEndpointsSwitchoverV1Api.DeleteSwitchoverV1SwitchoverEndpoint(c.switchoverApiContext(), id).Environment(environment).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}
