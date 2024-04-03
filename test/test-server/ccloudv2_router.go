package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

type CloudV2Router struct {
	*mux.Router
}

var ccloudV2Routes = []route{
	{"/byok/v1/keys", handleByokKeys},
	{"/byok/v1/keys/{id}", handleByokKey},
	{"/billing/v1/costs", handleBillingCosts},
	{"/cdx/v1/consumer-shared-resources", handleConsumerSharedResources},
	{"/cdx/v1/consumer-shared-resources/{id}:network", handlePrivateLinkNetworkConfig},
	{"/cdx/v1/consumer-shares", handleStreamSharingConsumerShares},
	{"/cdx/v1/consumer-shares/{id}", handleStreamSharingConsumerShare},
	{"/cdx/v1/opt-in", handleOptInOptOut},
	{"/cdx/v1/provider-shares", handleStreamSharingProviderShares},
	{"/cdx/v1/provider-shares/{id}", handleStreamSharingProviderShare},
	{"/cdx/v1/provider-shares/{id}:resend", handleStreamSharingResendInvite},
	{"/cdx/v1/shared-tokens:redeem", handleStreamSharingRedeemToken},
	{"/cmk/v2/clusters", handleCmkClusters},
	{"/cmk/v2/clusters/{id}", handleCmkCluster},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connector-plugins", handlePlugins},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connector-plugins/{plugin}/config/validate", handlePluginValidate},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connectors", handleConnectors},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}", handleConnector},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/config", handleConnectorConfig},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/pause", handleConnectorPause},
	{"/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/resume", handleConnectorResume},
	{"/connect/v1/environments/{environment_id}/clusters/{kafka_cluster_id}/connectors/{connector_name}/offsets", handleConnectorOffsets},
	{"/connect/v1/environments/{environment_id}/clusters/{kafka_cluster_id}/connectors/{connector_name}/offsets/request", handleAlterConnectorOffsets},
	{"/connect/v1/environments/{environment_id}/clusters/{kafka_cluster_id}/connectors/{connector_name}/offsets/request/status", handleAlterConnectorOffsetsStatus},
	{"/connect/v1/custom-connector-plugins", handleCustomConnectorPlugins},
	{"/connect/v1/custom-connector-plugins/{id}", handleCustomConnectorPluginsId},
	{"/connect/v1/presigned-upload-url", handleCustomPluginUploadUrl},
	{"/connect/v1/dummy-presigned-url", handleCustomPluginUploadFile},
	{"/fcpm/v2/compute-pools", handleFcpmComputePools},
	{"/fcpm/v2/compute-pools/{id}", handleFcpmComputePoolsId},
	{"/fcpm/v2/regions", handleFcpmRegions},
	{"/iam/v2/api-keys", handleIamApiKeys},
	{"/iam/v2/api-keys/{id}", handleIamApiKey},
	{"/iam/v2/identity-providers", handleIamIdentityProviders},
	{"/iam/v2/identity-providers/{id}", handleIamIdentityProvider},
	{"/iam/v2/identity-providers/{provider_id}/identity-pools", handleIamIdentityPools},
	{"/iam/v2/identity-providers/{provider_id}/identity-pools/{id}", handleIamIdentityPool},
	{"/iam/v2/ip-filters", handleIamIpFilters},
	{"/iam/v2/ip-filters/{id}", handleIamIpFilter},
	{"/iam/v2/ip-groups", handleIamIpGroups},
	{"/iam/v2/ip-groups/{id}", handleIamIpGroup},
	{"/iam/v2/invitations", handleIamInvitations},
	{"/iam/v2/role-bindings", handleIamRoleBindings},
	{"/iam/v2/role-bindings/{id}", handleIamRoleBinding},
	{"/iam/v2/service-accounts", handleIamServiceAccounts},
	{"/iam/v2/service-accounts/{id}", handleIamServiceAccount},
	{"/iam/v2/sso/group-mappings", handleIamGroupMappings},
	{"/iam/v2/sso/group-mappings/{id}", handleIamGroupMapping},
	{"/iam/v2/users", handleIamUsers},
	{"/iam/v2/users/{id}", handleIamUser},
	{"/kafka-quotas/v1/client-quotas", handleKafkaClientQuotas},
	{"/kafka-quotas/v1/client-quotas/{id}", handleKafkaClientQuota},
	{"/ksqldbcm/v2/clusters", handleKsqlClusters},
	{"/ksqldbcm/v2/clusters/{id}", handleKsqlCluster},
	{"/networking/v1/access-points", handleNetworkingAccessPoints},
	{"/networking/v1/access-points/{id}", handleNetworkingAccessPoint},
	{"/networking/v1/dns-forwarders", handleNetworkingDnsForwarders},
	{"/networking/v1/dns-forwarders/{id}", handleNetworkingDnsForwarder},
	{"/networking/v1/dns-records", handleNetworkingDnsRecords},
	{"/networking/v1/dns-records/{id}", handleNetworkingDnsRecord},
	{"/networking/v1/gateways", handleNetworkingGateways},
	{"/networking/v1/gateways/{id}", handleNetworkingGateway},
	{"/networking/v1/ip-addresses", handleNetworkingIpAddresses},
	{"/networking/v1/networks", handleNetworkingNetworks},
	{"/networking/v1/networks/{id}", handleNetworkingNetwork},
	{"/networking/v1/peerings", handleNetworkingPeerings},
	{"/networking/v1/peerings/{id}", handleNetworkingPeering},
	{"/networking/v1/private-link-accesses", handleNetworkingPrivateLinkAccesses},
	{"/networking/v1/private-link-accesses/{id}", handleNetworkingPrivateLinkAccess},
	{"/networking/v1/private-link-attachments", handleNetworkingPrivateLinkAttachments},
	{"/networking/v1/private-link-attachments/{id}", handleNetworkingPrivateLinkAttachment},
	{"/networking/v1/private-link-attachment-connections", handleNetworkingPrivateLinkAttachmentConnections},
	{"/networking/v1/private-link-attachment-connections/{id}", handleNetworkingPrivateLinkAttachmentConnection},
	{"/networking/v1/transit-gateway-attachments", handleNetworkingTransitGatewayAttachments},
	{"/networking/v1/transit-gateway-attachments/{id}", handleNetworkingTransitGatewayAttachment},
	{"/org/v2/environments", handleOrgEnvironments},
	{"/org/v2/environments/{id}", handleOrgEnvironment},
	{"/org/v2/organizations", handleOrgOrganizations},
	{"/org/v2/organizations/{id}", handleOrgOrganization},
	{"/sd/v1/pipelines", handlePipelines},
	{"/sd/v1/pipelines/{id}", handlePipeline},
	{"/service-quota/v1/applied-quotas", handleAppliedQuotas},
	{"/service-quota/v2/applied-quotas", handleAppliedQuotas},
	{"/srcm/v2/clusters", handleSchemaRegistryClusters},
	{"/srcm/v2/clusters/{id}", handleSchemaRegistryCluster},
	{"/srcm/v2/regions", handleSchemaRegistryRegions},
	{"/srcm/v2/regions/{id}", handleSchemaRegistryRegion},
	{"/v2/metrics/cloud/query", handleMetricsQuery},
}

func NewV2Router(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range ccloudV2Routes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}
