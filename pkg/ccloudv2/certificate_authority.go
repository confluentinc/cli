package ccloudv2

import (
	"context"
	"net/http"

	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newCertificateAuthorityClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *certificateauthorityv2.APIClient {
	cfg := certificateauthorityv2.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = certificateauthorityv2.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return certificateauthorityv2.NewAPIClient(cfg)
}

func (c *Client) certificateAuthorityApiContext() context.Context {
	return context.WithValue(context.Background(), certificateauthorityv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) certificatePoolApiContext() context.Context {
	return context.WithValue(context.Background(), certificateauthorityv2.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateCertificateAuthority(certRequest certificateauthorityv2.IamV2CreateCertRequest) (certificateauthorityv2.IamV2CertificateAuthority, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateAuthoritiesIamV2Api.CreateIamV2CertificateAuthority(c.certificateAuthorityApiContext()).IamV2CreateCertRequest(certRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetCertificateAuthority(id string) (certificateauthorityv2.IamV2CertificateAuthority, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateAuthoritiesIamV2Api.GetIamV2CertificateAuthority(c.certificateAuthorityApiContext(), id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCertificateAuthority(certRequest certificateauthorityv2.IamV2UpdateCertRequest) (certificateauthorityv2.IamV2CertificateAuthority, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateAuthoritiesIamV2Api.UpdateIamV2CertificateAuthority(c.certificateAuthorityApiContext(), certRequest.GetId()).IamV2UpdateCertRequest(certRequest).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCertificateAuthority(id string) error {
	_, httpResp, err := c.CertificateAuthorityClient.CertificateAuthoritiesIamV2Api.DeleteIamV2CertificateAuthority(c.certificateAuthorityApiContext(), id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCertificateAuthorities() ([]certificateauthorityv2.IamV2CertificateAuthority, error) {
	var list []certificateauthorityv2.IamV2CertificateAuthority

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCertificateAuthorities(pageToken)
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

func (c *Client) executeListCertificateAuthorities(pageToken string) (certificateauthorityv2.IamV2CertificateAuthorityList, *http.Response, error) {
	req := c.CertificateAuthorityClient.CertificateAuthoritiesIamV2Api.ListIamV2CertificateAuthorities(c.certificateAuthorityApiContext()).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) CreateCertificatePool(certificatePool certificateauthorityv2.IamV2CertificateIdentityPool, provider string) (certificateauthorityv2.IamV2CertificateIdentityPool, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateIdentityPoolsIamV2Api.CreateIamV2CertificateIdentityPool(c.certificatePoolApiContext(), provider).IamV2CertificateIdentityPool(certificatePool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetCertificatePool(id string, provider string) (certificateauthorityv2.IamV2CertificateIdentityPool, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateIdentityPoolsIamV2Api.GetIamV2CertificateIdentityPool(c.certificatePoolApiContext(), provider, id).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCertificatePool(certificatePool certificateauthorityv2.IamV2CertificateIdentityPool, provider string) (certificateauthorityv2.IamV2CertificateIdentityPool, error) {
	resp, httpResp, err := c.CertificateAuthorityClient.CertificateIdentityPoolsIamV2Api.UpdateIamV2CertificateIdentityPool(c.certificatePoolApiContext(), provider, *certificatePool.Id).IamV2CertificateIdentityPool(certificatePool).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCertificatePool(id string, provider string) error {
	_, httpResp, err := c.CertificateAuthorityClient.CertificateIdentityPoolsIamV2Api.DeleteIamV2CertificateIdentityPool(c.certificatePoolApiContext(), provider, id).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListCertificatePool(providerID string) ([]certificateauthorityv2.IamV2CertificateIdentityPool, error) {
	var list []certificateauthorityv2.IamV2CertificateIdentityPool

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCertificatePool(providerID, pageToken)
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

func (c *Client) executeListCertificatePool(providerID, pageToken string) (certificateauthorityv2.IamV2CertificateIdentityPoolList, *http.Response, error) {
	req := c.CertificateAuthorityClient.CertificateIdentityPoolsIamV2Api.ListIamV2CertificateIdentityPools(c.certificatePoolApiContext(), providerID).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}
