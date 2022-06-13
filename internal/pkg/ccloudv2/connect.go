package ccloudv2

import (
	"context"
	"net/http"

	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	plog "github.com/confluentinc/cli/internal/pkg/log"
)

func newConnectClient(baseURL, userAgent string, isTest bool) *connectv1.APIClient {
	cfg := connectv1.NewConfiguration()
	cfg.Debug = plog.CliLogger.Level >= plog.DEBUG
	cfg.HTTPClient = newRetryableHttpClient()
	cfg.Servers = connectv1.ServerConfigurations{{URL: getServerUrl(baseURL, isTest), Description: "Confluent Cloud Connect"}}
	cfg.UserAgent = userAgent

	return connectv1.NewAPIClient(cfg)
}

func (c *Client) connectApiContext() context.Context {
	return context.WithValue(context.Background(), connectv1.ContextAccessToken, c.AuthToken)
}

func (c *Client) CreateConnector(environmentId, kafkaClusterId string, connect connectv1.InlineObject) (connectv1.ConnectV1Connector, *http.Response, error) {
	req := c.ConnectClient.ConnectorsV1Api.CreateConnectv1Connector(c.connectApiContext(), environmentId, kafkaClusterId).InlineObject(connect)
	return c.ConnectClient.ConnectorsV1Api.CreateConnectv1ConnectorExecute(req)
}

func (c *Client) CreateOrUpdateConnectorConfig(connectorName, environmentId, kafkaClusterId string, configs map[string]string) (connectv1.ConnectV1Connector, *http.Response, error) {
	req := c.ConnectClient.ConnectorsV1Api.CreateOrUpdateConnectv1ConnectorConfig(c.connectApiContext(), connectorName, environmentId, kafkaClusterId).RequestBody(configs)
	return c.ConnectClient.ConnectorsV1Api.CreateOrUpdateConnectv1ConnectorConfigExecute(req)
}

func (c *Client) DeleteConnector(connectorName, environmentId, kafkaClusterId string) (connectv1.InlineResponse200, *http.Response, error) {
	req := c.ConnectClient.ConnectorsV1Api.DeleteConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId)
	return c.ConnectClient.ConnectorsV1Api.DeleteConnectv1ConnectorExecute(req)
}

func (c *Client) ListConnectorsWithExpansions(environmentId, kafkaClusterId, expand string) (map[string]connectv1.ConnectV1ConnectorExpansion, *http.Response, error) {
	req := c.ConnectClient.ConnectorsV1Api.ListConnectv1ConnectorsWithExpansions(c.connectApiContext(), environmentId, kafkaClusterId).Expand(expand)
	return c.ConnectClient.ConnectorsV1Api.ListConnectv1ConnectorsWithExpansionsExecute(req)
}

func (c *Client) GetConnectorExpansionById(connectorId, environmentId, kafkaClusterId string) (*connectv1.ConnectV1ConnectorExpansion, error) {
	connectorExpansions, _, err := c.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "status,info,id")
	if err != nil {
		return nil, err
	}

	for _, connector := range connectorExpansions {
		if connector.Id.GetId() == connectorId {
			return &connector, nil
		}
	}

	return nil, errors.Errorf(errors.UnknownConnectorIdErrorMsg, connectorId)
}

func (c *Client) GetConnectorExpansionByName(connectorName, environmentId, kafkaClusterId string) (*connectv1.ConnectV1ConnectorExpansion, error) {
	connectorExpansions, _, err := c.ListConnectorsWithExpansions(environmentId, kafkaClusterId, "status,info,id")
	if err != nil {
		return nil, err
	}

	for name, connector := range connectorExpansions {
		if name == connectorName {
			return &connector, nil
		}
	}

	return nil, errors.Errorf(errors.UnknownConnectorIdErrorMsg, connectorName)
}

func (c *Client) PauseConnector(connectorName, environmentId, kafkaClusterId string) (*http.Response, error) {
	req := c.ConnectClient.LifecycleV1Api.PauseConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId)
	return c.ConnectClient.LifecycleV1Api.PauseConnectv1ConnectorExecute(req)
}

func (c *Client) ResumeConnector(connectorName, environmentId, kafkaClusterId string) (*http.Response, error) {
	req := c.ConnectClient.LifecycleV1Api.ResumeConnectv1Connector(c.connectApiContext(), connectorName, environmentId, kafkaClusterId)
	return c.ConnectClient.LifecycleV1Api.ResumeConnectv1ConnectorExecute(req)
}

func (c *Client) ListConnectorPlugins(environmentId, kafkaClusterId string) ([]connectv1.InlineResponse2002, *http.Response, error) {
	req := c.ConnectClient.PluginsV1Api.ListConnectv1ConnectorPlugins(c.connectApiContext(), environmentId, kafkaClusterId)
	return c.ConnectClient.PluginsV1Api.ListConnectv1ConnectorPluginsExecute(req)
}

func (c *Client) ValidateConnectorPlugin(pluginName, environmentId, kafkaClusterId string, configs map[string]string) (connectv1.InlineResponse2003, *http.Response, error) {
	req := c.ConnectClient.PluginsV1Api.ValidateConnectv1ConnectorPlugin(c.connectApiContext(), pluginName, environmentId, kafkaClusterId).RequestBody(configs)
	return c.ConnectClient.PluginsV1Api.ValidateConnectv1ConnectorPluginExecute(req)
}
