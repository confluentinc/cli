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
	{"/fcpm/v2/compute-pools", handleFcpmComputePools},
	{"/fcpm/v2/compute-pools/{id}", handleFcpmComputePoolsId},
	{"/fcpm/v2/iam-bindings", handleFcpmIamBindings},
	{"/fcpm/v2/iam-bindings/{id}", handleFcpmIamBindingsId},
	{"/fcpm/v2/regions", handleFcpmRegions},
	{"/iam/v2/api-keys", handleIamApiKeys},
	{"/iam/v2/api-keys/{id}", handleIamApiKey},
	{"/iam/v2/identity-providers", handleIamIdentityProviders},
	{"/iam/v2/identity-providers/{id}", handleIamIdentityProvider},
	{"/iam/v2/identity-providers/{provider_id}/identity-pools", handleIamIdentityPools},
	{"/iam/v2/identity-providers/{provider_id}/identity-pools/{id}", handleIamIdentityPool},
	{"/iam/v2/invitations", handleIamInvitations},
	{"/iam/v2/role-bindings", handleIamRoleBindings},
	{"/iam/v2/role-bindings/{id}", handleIamRoleBinding},
	{"/iam/v2/service-accounts", handleIamServiceAccounts},
	{"/iam/v2/service-accounts/{id}", handleIamServiceAccount},
	{"/iam/v2/users", handleIamUsers},
	{"/iam/v2/users/{id}", handleIamUser},
	{"/kafka-quotas/v1/client-quotas", handleKafkaClientQuotas},
	{"/kafka-quotas/v1/client-quotas/{id}", handleKafkaClientQuota},
	{"/ksqldbcm/v2/clusters", handleKsqlClusters},
	{"/ksqldbcm/v2/clusters/{id}", handleKsqlCluster},
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
