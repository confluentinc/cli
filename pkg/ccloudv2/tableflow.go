package ccloudv2

import (
	"context"
	"net/http"

	tableflowv1 "github.com/confluentinc/ccloud-sdk-go-v2/tableflow/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newTableflowClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *tableflowv1.APIClient {
	cfg := tableflowv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = tableflowv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return tableflowv1.NewAPIClient(cfg)
}

func (c *Client) tableflowApiContext() context.Context {
	return context.WithValue(context.Background(), tableflowv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

// TableflowTopicsTableflowV1Api
func (c *Client) ListTableflowTopics(environment string, cluster string) ([]tableflowv1.TableflowV1TableflowTopic, error) {
	var list []tableflowv1.TableflowV1TableflowTopic

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListTableflowTopics(environment, cluster, pageToken)
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

func (c *Client) executeListTableflowTopics(environment, cluster, pageToken string) (tableflowv1.TableflowV1TableflowTopicList, *http.Response, error) {
	req := c.TableflowClient.TableflowTopicsTableflowV1Api.ListTableflowV1TableflowTopics(c.tableflowApiContext()).SpecKafkaCluster(cluster).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) GetTableflowTopic(environment, cluster, display_name string) (tableflowv1.TableflowV1TableflowTopic, error) {
	resp, httpResp, err := c.TableflowClient.TableflowTopicsTableflowV1Api.GetTableflowV1TableflowTopic(c.tableflowApiContext(), display_name).SpecKafkaCluster(cluster).Environment(environment).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteTableflowTopic(environment, cluster, display_name string) error {
	httpResp, err := c.TableflowClient.TableflowTopicsTableflowV1Api.DeleteTableflowV1TableflowTopic(c.tableflowApiContext(), display_name).SpecKafkaCluster(cluster).Environment(environment).Execute()

	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateTableflowTopic(display_name string, topicUpdate tableflowv1.TableflowV1TableflowTopicUpdate) (tableflowv1.TableflowV1TableflowTopic, error) {
	resp, httpResp, err := c.TableflowClient.TableflowTopicsTableflowV1Api.UpdateTableflowV1TableflowTopic(c.tableflowApiContext(), display_name).TableflowV1TableflowTopicUpdate(topicUpdate).Execute()

	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateTableflowTopic(topic tableflowv1.TableflowV1TableflowTopic) (tableflowv1.TableflowV1TableflowTopic, error) {
	resp, httpResp, err := c.TableflowClient.TableflowTopicsTableflowV1Api.CreateTableflowV1TableflowTopic(c.tableflowApiContext()).TableflowV1TableflowTopic(topic).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

// CatalogIntegrationTableflowV1Api
func (c *Client) ListCatalogIntegrations(environment string, cluster string) ([]tableflowv1.TableflowV1CatalogIntegration, error) {
	var list []tableflowv1.TableflowV1CatalogIntegration

	done := false
	pageToken := ""
	for !done {
		page, httpResp, err := c.executeListCatalogIntegration(environment, cluster, pageToken)
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

func (c *Client) executeListCatalogIntegration(environment, cluster, pageToken string) (tableflowv1.TableflowV1CatalogIntegrationList, *http.Response, error) {
	req := c.TableflowClient.CatalogIntegrationsTableflowV1Api.ListTableflowV1CatalogIntegrations(c.tableflowApiContext()).SpecKafkaCluster(cluster).Environment(environment).PageSize(ccloudV2ListPageSize)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	return req.Execute()
}

func (c *Client) GetCatalogIntegration(environment, cluster, display_name string) (tableflowv1.TableflowV1CatalogIntegration, error) {
	resp, httpResp, err := c.TableflowClient.CatalogIntegrationsTableflowV1Api.GetTableflowV1CatalogIntegration(c.tableflowApiContext(), display_name).SpecKafkaCluster(cluster).Environment(environment).Execute()

	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteCatalogIntegration(environment, cluster, display_name string) error {
	httpResp, err := c.TableflowClient.CatalogIntegrationsTableflowV1Api.DeleteTableflowV1CatalogIntegration(c.tableflowApiContext(), display_name).SpecKafkaCluster(cluster).Environment(environment).Execute()

	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) UpdateCatalogIntegration(display_name string, catalogIntegrationUpdate tableflowv1.TableflowV1CatalogIntegrationUpdateRequest) (tableflowv1.TableflowV1CatalogIntegration, error) {
	resp, httpResp, err := c.TableflowClient.CatalogIntegrationsTableflowV1Api.UpdateTableflowV1CatalogIntegration(c.tableflowApiContext(), display_name).TableflowV1CatalogIntegrationUpdateRequest(catalogIntegrationUpdate).Execute()

	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateCatalogIntegration(catalogIntegration tableflowv1.TableflowV1CatalogIntegration) (tableflowv1.TableflowV1CatalogIntegration, error) {
	resp, httpResp, err := c.TableflowClient.CatalogIntegrationsTableflowV1Api.CreateTableflowV1CatalogIntegration(c.tableflowApiContext()).TableflowV1CatalogIntegration(catalogIntegration).Execute()

	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
