package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
)

func (c *CloudRouter) HandleAllRolesRoute(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		roles := rbacPublicRoles()
		rolesResponse, _ := json.Marshal(roles)
		_, err := w.Write(rolesResponse)
		require.NoError(t, err)
	}
}

func rbacPublicRoles() []mdsv2alpha1.Role {
	cloudClusterAdminRole := mdsv2alpha1.Role{
		Name: "CloudClusterAdmin",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "cluster",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Topic", Operations: []string{"All"}},
					{ResourceType: "KsqlCluster", Operations: []string{"All"}},
					{ResourceType: "Subject", Operations: []string{"All"}},
					{ResourceType: "Connector", Operations: []string{"All"}},
					{ResourceType: "NetworkAccess", Operations: []string{"All"}},
					{ResourceType: "ClusterMetric", Operations: []string{"All"}},
					{ResourceType: "Cluster", Operations: []string{"All"}},
					{ResourceType: "ClusterApiKey", Operations: []string{"All"}},
					{ResourceType: "SecurityMetadata", Operations: []string{"Describe", "Alter"}},
				},
			},
			{
				BindingScope: "organization",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "SupportPlan", Operations: []string{"Describe"}},
					{ResourceType: "User", Operations: []string{"Describe", "Invite"}},
					{ResourceType: "ServiceAccount", Operations: []string{"Describe"}},
				},
			},
		},
	}

	environmentAdminRole := mdsv2alpha1.Role{
		Name: "EnvironmentAdmin",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "ENVIRONMENT",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "SecurityMetadata", Operations: []string{"Describe", "Alter"}},
					{ResourceType: "ClusterApiKey", Operations: []string{"All"}},
					{ResourceType: "Connector", Operations: []string{"All"}},
					{ResourceType: "NetworkAccess", Operations: []string{"All"}},
					{ResourceType: "KsqlCluster", Operations: []string{"All"}},
					{ResourceType: "Environment", Operations: []string{"Alter", "Delete", "AlterAccess", "CreateKafkaCluster", "DescribeAccess"}},
					{ResourceType: "Subject", Operations: []string{"All"}},
					{ResourceType: "NetworkConfig", Operations: []string{"All"}},
					{ResourceType: "ClusterMetric", Operations: []string{"All"}},
					{ResourceType: "Cluster", Operations: []string{"All"}},
					{ResourceType: "SchemaRegistry", Operations: []string{"All"}},
					{ResourceType: "NetworkRegion", Operations: []string{"All"}},
					{ResourceType: "Deployment", Operations: []string{"All"}},
					{ResourceType: "Topic", Operations: []string{"All"}},
				},
			},
			{
				BindingScope: "organization",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "User", Operations: []string{"Describe", "Invite"}},
					{ResourceType: "ServiceAccount", Operations: []string{"Describe"}},
					{ResourceType: "SupportPlan", Operations: []string{"Describe"}},
				},
			},
		},
	}

	organizationAdminRole := mdsv2alpha1.Role{
		Name: "OrganizationAdmin",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "organization",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Topic", Operations: []string{"All"}},
					{ResourceType: "NetworkConfig", Operations: []string{"All"}},
					{ResourceType: "SecurityMetadata", Operations: []string{"Describe", "Alter"}},
					{ResourceType: "Billing", Operations: []string{"All"}},
					{ResourceType: "ClusterApiKey", Operations: []string{"All"}},
					{ResourceType: "Deployment", Operations: []string{"All"}},
					{ResourceType: "SchemaRegistry", Operations: []string{"All"}},
					{ResourceType: "KsqlCluster", Operations: []string{"All"}},
					{ResourceType: "CloudApiKey", Operations: []string{"All"}},
					{ResourceType: "NetworkAccess", Operations: []string{"All"}},
					{ResourceType: "SecuritySSO", Operations: []string{"All"}},
					{ResourceType: "SupportPlan", Operations: []string{"All"}},
					{ResourceType: "Connector", Operations: []string{"All"}},
					{ResourceType: "ClusterMetric", Operations: []string{"All"}},
					{ResourceType: "ServiceAccount", Operations: []string{"All"}},
					{ResourceType: "Subject", Operations: []string{"All"}},
					{ResourceType: "Cluster", Operations: []string{"All"}},
					{ResourceType: "Environment", Operations: []string{"All"}},
					{ResourceType: "NetworkRegion", Operations: []string{"All"}},
					{ResourceType: "Organization", Operations: []string{"Alter", "CreateEnvironment", "AlterAccess", "DescribeAccess"}},
					{ResourceType: "User", Operations: []string{"All"}},
				},
			},
		},
	}

	resourceOwnerRole := mdsv2alpha1.Role{
		Name: "ResourceOwner",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "cloud-cluster",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "CloudCluster", Operations: []string{"Describe"}},
				},
			},
			{
				BindingScope:     "cluster",
				BindWithResource: true,
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Topic", Operations: []string{"Create", "Delete", "Read", "Write", "Describe", "DescribeConfigs", "Alter", "AlterConfigs", "DescribeAccess", "AlterAccess"}},
					{ResourceType: "Group", Operations: []string{"Read", "Describe", "Delete", "DescribeAccess", "AlterAccess"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{cloudClusterAdminRole, environmentAdminRole, organizationAdminRole, resourceOwnerRole}
}

func rolesListToJsonMap(roles []mdsv2alpha1.Role) map[string]string {
	roleMap := make(map[string]string)
	for _, role := range roles {
		jsonVal, _ := json.Marshal(role)
		roleMap[role.Name] = string(jsonVal)
	}

	return roleMap
}
