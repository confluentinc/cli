package ccloudv2

import (
	"context"
	"fmt"
	"net/http"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

func newConnectClient(httpClient *http.Client, url, userAgent string, unsafeTrace bool) *connectv1.APIClient {
	cfg := connectv1.NewConfiguration()
	cfg.Debug = unsafeTrace
	cfg.HTTPClient = httpClient
	cfg.Servers = connectv1.ServerConfigurations{{URL: url}}
	cfg.UserAgent = userAgent

	return connectv1.NewAPIClient(cfg)
}

func (c *Client) connectApiContext() context.Context {
	return context.WithValue(context.Background(), connectv1.ContextAccessToken, c.cfg.Context().GetAuthToken())
}

func (c *Client) CreateConnector(environmentId, kafkaClusterId string, connect connectv1.InlineObject) (connectv1.ConnectV1ConnectorWithOffsets, error) {
	resp, httpResp, err := c.ConnectClient.ConnectorsConnectV1Api.CreateConnectv1Connector(c.connectApiContext(), environmentId, kafkaClusterId).InlineObject(connect).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) CreateOrUpdateConnectorConfig(connectorName, environmentId, kafkaClusterId string, configs map[string]string) (connectv1.ConnectV1Connector, error) {
	resp, httpResp, err := c.ConnectClient.ConnectorsConnectV1Api.CreateOrUpdateConnectv1ConnectorConfig(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).RequestBody(configs).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) DeleteConnector(connectorName, environmentId, kafkaClusterId string) (connectv1.InlineResponse200, error) {
	resp, httpResp, err := c.ConnectClient.ConnectorsConnectV1Api.DeleteConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListConnectorsWithExpansions(environmentId, kafkaClusterId, expand string) (map[string]connectv1.ConnectV1ConnectorExpansion, error) {
	resp, httpResp, err := c.ConnectClient.ConnectorsConnectV1Api.ListConnectv1ConnectorsWithExpansions(c.connectApiContext(), environmentId, kafkaClusterId).Expand(expand).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetConnectorExpansionById(connectorId, environmentId, kafkaClusterId string) (*connectv1.ConnectV1ConnectorExpansion, error) {
	connectors, err := c.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "id,info,status")
	if err != nil {
		return nil, err
	}

	for _, connector := range connectors {
		if connector.Id.GetId() == connectorId {
			return &connector, nil
		}
	}

	return nil, fmt.Errorf(errors.UnknownConnectorIdErrorMsg, connectorId)
}

func (c *Client) GetConnectorExpansionByName(connectorName, environmentId, kafkaClusterId string) (*connectv1.ConnectV1ConnectorExpansion, error) {
	connectors, err := c.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "id,status")
	if err != nil {
		return nil, err
	}

	for name, connector := range connectors {
		if name == connectorName {
			return &connector, nil
		}
	}

	return nil, fmt.Errorf(errors.UnknownConnectorIdErrorMsg, connectorName)
}

func (c *Client) PauseConnector(connectorName, environmentId, kafkaClusterId string) error {
	httpResp, err := c.ConnectClient.LifecycleConnectV1Api.PauseConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ResumeConnector(connectorName, environmentId, kafkaClusterId string) error {
	httpResp, err := c.ConnectClient.LifecycleConnectV1Api.ResumeConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).Execute()
	return errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ListConnectorPlugins(environmentId, kafkaClusterId string) ([]connectv1.InlineResponse2002, error) {
	resp, httpResp, err := c.ConnectClient.ManagedConnectorPluginsConnectV1Api.ListConnectv1ConnectorPlugins(c.connectApiContext(), environmentId, kafkaClusterId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) ValidateConnectorPlugin(pluginName, environmentId, kafkaClusterId string, configs map[string]string) (connectv1.InlineResponse2003, error) {
	resp, httpResp, err := c.ConnectClient.ManagedConnectorPluginsConnectV1Api.ValidateConnectv1ConnectorPlugin(c.connectApiContext(), pluginName, environmentId, kafkaClusterId).RequestBody(configs).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) GetConnectorOffset(connectorName, environmentId, kafkaClusterId string) (connectv1.ConnectV1ConnectorOffsets, error) {
	offsets, httpResp, err := c.ConnectClient.OffsetsConnectV1Api.GetConnectv1ConnectorOffsets(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).Execute()
	return offsets, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) AlterConnectorOffsets(connectorName, environmentId, kafkaClusterId string, requestBody connectv1.ConnectV1AlterOffsetRequest) (connectv1.ConnectV1AlterOffsetRequestInfo, error) {
	resp, httpResp, err := c.ConnectClient.OffsetsConnectV1Api.AlterConnectv1ConnectorOffsetsRequest(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).ConnectV1AlterOffsetRequest(requestBody).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}

func (c *Client) AlterConnectorOffsetsRequestStatus(pluginName, environmentId, kafkaClusterId string) (connectv1.ConnectV1AlterOffsetStatus, error) {
	resp, httpResp, err := c.ConnectClient.OffsetsConnectV1Api.GetConnectv1ConnectorOffsetsRequestStatus(c.connectApiContext(), pluginName, environmentId, kafkaClusterId).Execute()
	return resp, errors.CatchCCloudV2Error(err, httpResp)
}
