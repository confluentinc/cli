package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/confluentinc/mds-sdk-go-public/mdsv2alpha1"
	"github.com/stretchr/testify/require"
)

func (c *CloudRouter) HandleAllRolesRoute(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := r.URL.Query().Get("namespace")
		namespaces := strings.Split(namespace, ",")

		var allRoles []mdsv2alpha1.Role
		for _, ns := range namespaces {
			switch ns {
			case "ksql":
				allRoles = append(allRoles, rbacKsqlRoles()...)
			case "datagovernance":
				allRoles = append(allRoles, rbacSRRoles()...)
			case "dataplane":
				allRoles = append(allRoles, rbacDataPlaneRoles()...)
			case "public":
				allRoles = append(allRoles, rbacPublicRoles()...)
			case "streamcatalog":
				allRoles = append(allRoles, rbacStreamCatalogRoles()...)
			default:
				allRoles = append(allRoles, rbacPublicRoles()...)
			}
		}

		allRolesResponse, _ := json.Marshal(allRoles)
		_, err := w.Write(allRolesResponse)
		require.NoError(t, err)
	}
}

func rbacDataPlaneRoles() []mdsv2alpha1.Role {
	developerManageRole := mdsv2alpha1.Role{
		Name: "DeveloperManage",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "cloud-cluster",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "CloudCluster", Operations: []string{"Describe"}},
				},
			},
			{
				BindingScope: "cluster",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Cluster", Operations: []string{"View", "AccessWithToken"}},
					{ResourceType: "OwnKafkaClusterApiKey", Operations: []string{"Describe", "Alter", "Delete", "Create"}},
					{ResourceType: "OwnClusterApiKey", Operations: []string{"Describe", "Alter", "Delete", "Create"}},
				},
			},
			{
				BindingScope:     "cluster",
				BindWithResource: true,
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Topic", Operations: []string{"Delete", "Describe", "Create", "DescribeConfigs"}},
					{ResourceType: "Cluster", Operations: []string{"Create", "DescribeConfigs"}},
					{ResourceType: "TransactionalId", Operations: []string{"Describe"}},
					{ResourceType: "Group", Operations: []string{"Describe", "Delete"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{developerManageRole}
}

func rbacKsqlRoles() []mdsv2alpha1.Role {
	resourceOwnerRole := mdsv2alpha1.Role{
		Name: "ResourceOwner",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope:     "ksql-cluster",
				BindWithResource: true,
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "KsqlCluster", Operations: []string{"Describe", "AlterAccess", "Contribute", "DescribeAccess", "Terminate"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{resourceOwnerRole}
}

func rbacSRRoles() []mdsv2alpha1.Role {
	resourceOwnerRole := mdsv2alpha1.Role{
		Name: "ResourceOwner",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope:     "schema-registry-cluster",
				BindWithResource: true,
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "Subject", Operations: []string{"Delete", "Read", "Write", "ReadCompatibility", "AlterAccess", "WriteCompatibility", "DescribeAccess"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{resourceOwnerRole}
}

func rbacPublicRoles() []mdsv2alpha1.Role {
	ccloudRoleBindingAdminRole := mdsv2alpha1.Role{
		Name: "CCloudRoleBindingAdmin",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope: "root",
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "SecurityMetadata", Operations: []string{"Describe", "Alter"}},
					{ResourceType: "Organization", Operations: []string{"AlterAccess", "DescribeAccess"}},
				},
			},
		},
	}

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

	return []mdsv2alpha1.Role{ccloudRoleBindingAdminRole, cloudClusterAdminRole, environmentAdminRole, organizationAdminRole, resourceOwnerRole}
}

func rbacStreamCatalogRoles() []mdsv2alpha1.Role {
	dataDiscoveryRole := mdsv2alpha1.Role{
		Name: "DataDiscovery",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope:     "environment",
				BindWithResource: false,
				AllowedOperations: []mdsv2alpha1.Operation{
					{ResourceType: "CatalogTagDefinition", Operations: []string{"Read"}},
					{ResourceType: "Topic", Operations: []string{"ReadCatalog"}},
					{ResourceType: "Subject", Operations: []string{"Read", "ReadCatalog", "ReadCompatibility"}},
					{ResourceType: "CatalogBusinessMetadataDefinition", Operations: []string{"Read"}},
				},
			},
		},
	}

	dataStewardRole := mdsv2alpha1.Role{
		Name: "DataSteward",
		Policies: []mdsv2alpha1.AccessPolicy{
			{
				BindingScope:     "environment",
				BindWithResource: false,
				AllowedOperations: []mdsv2alpha1.Operation{
					{
						ResourceType: "CatalogTagDefinition",
						Operations:   []string{"Read", "Write", "Delete"},
					},
					{
						ResourceType: "Topic",
						Operations:   []string{"ReadCatalog", "WriteCatalog"},
					},
					{
						ResourceType: "Subject",
						Operations:   []string{"Delete", "Read", "ReadCatalog", "ReadCompatibility", "Write", "WriteCatalog", "WriteCompatibility"},
					},
					{
						ResourceType: "CatalogBusinessMetadataDefinition",
						Operations:   []string{"Read", "Write", "Delete"},
					},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{dataDiscoveryRole, dataStewardRole}
}

func rolesListToJsonMap(roles []mdsv2alpha1.Role) map[string]string {
	roleMap := make(map[string]string)
	for _, role := range roles {
		jsonVal, _ := json.Marshal(role)
		roleMap[role.Name] = string(jsonVal)
	}

	return roleMap
}
