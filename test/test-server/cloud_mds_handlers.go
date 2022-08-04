package testserver

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/stretchr/testify/require"
)

func (c *CloudRouter) HandleAllRolesRoute(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		namespace := r.URL.Query().Get("namespace")
		namespaces := strings.Split(namespace, ",")

		var allRoles []mdsv2alpha1.Role
		for _, ns := range namespaces {
			switch ns {
			case "ksql":
				allRoles = append(allRoles, rbacKsqlRoles()...)
				break
			case "datagovernance":
				allRoles = append(allRoles, rbacSRRoles()...)
				break
			case "dataplane":
				allRoles = append(allRoles, rbacDataPlaneRoles()...)
				break
			case "public":
				allRoles = append(allRoles, rbacPublicRoles()...)
				break
			default:
				allRoles = append(allRoles, rbacPublicRoles()...)
				break
			}
		}

		allRolesResponse, _ := json.Marshal(allRoles)
		_, err := w.Write(allRolesResponse)
		require.NoError(t, err)
	}
}

func rbacDataPlaneRoles() []mdsv2alpha1.Role {
	developerManageRole := mdsv2alpha1.Role{
		"DeveloperManage",
		[]mdsv2alpha1.AccessPolicy{
			{
				"cloud-cluster",
				false,
				[]mdsv2alpha1.Operation{
					{"CloudCluster", []string{"Describe"}},
				},
			},
			{
				"cluster",
				false,
				[]mdsv2alpha1.Operation{
					{"Cluster", []string{"View", "AccessWithToken"}},
					{"OwnKafkaClusterApiKey", []string{"Describe", "Alter", "Delete", "Create"}},
					{"OwnClusterApiKey", []string{"Describe", "Alter", "Delete", "Create"}},
				},
			},
			{
				"cluster",
				true,
				[]mdsv2alpha1.Operation{
					{"Topic", []string{"Delete", "Describe", "Create", "DescribeConfigs"}},
					{"Cluster", []string{"Create", "DescribeConfigs"}},
					{"TransactionalId", []string{"Describe"}},
					{"Group", []string{"Describe", "Delete"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{developerManageRole}
}

func rbacKsqlRoles() []mdsv2alpha1.Role {
	resourceOwnerRole := mdsv2alpha1.Role{
		"ResourceOwner",
		[]mdsv2alpha1.AccessPolicy{
			{
				"ksql-cluster",
				true,
				[]mdsv2alpha1.Operation{
					{"KsqlCluster", []string{"Describe", "AlterAccess", "Contribute", "DescribeAccess", "Terminate"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{resourceOwnerRole}
}

func rbacSRRoles() []mdsv2alpha1.Role {
	resourceOwnerRole := mdsv2alpha1.Role{
		"ResourceOwner",
		[]mdsv2alpha1.AccessPolicy{
			{
				"schema-registry-cluster",
				true,
				[]mdsv2alpha1.Operation{
					{"Subject", []string{"Delete", "Read", "Write", "ReadCompatibility", "AlterAccess", "WriteCompatibility", "DescribeAccess"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{resourceOwnerRole}
}

func rbacPublicRoles() []mdsv2alpha1.Role {
	ccloudRoleBindingAdminRole := mdsv2alpha1.Role{
		"CCloudRoleBindingAdmin",
		[]mdsv2alpha1.AccessPolicy{
			{
				"root",
				false,
				[]mdsv2alpha1.Operation{
					{"SecurityMetadata", []string{"Describe", "Alter"}},
					{"Organization", []string{"AlterAccess", "DescribeAccess"}},
				},
			},
		},
	}

	cloudClusterAdminRole := mdsv2alpha1.Role{
		"CloudClusterAdmin",
		[]mdsv2alpha1.AccessPolicy{
			{
				"cluster",
				false,
				[]mdsv2alpha1.Operation{
					{"Topic", []string{"All"}},
					{"KsqlCluster", []string{"All"}},
					{"Subject", []string{"All"}},
					{"Connector", []string{"All"}},
					{"NetworkAccess", []string{"All"}},
					{"ClusterMetric", []string{"All"}},
					{"Cluster", []string{"All"}},
					{"ClusterApiKey", []string{"All"}},
					{"SecurityMetadata", []string{"Describe", "Alter"}},
				},
			},
			{
				"organization",
				false,
				[]mdsv2alpha1.Operation{
					{"SupportPlan", []string{"Describe"}},
					{"User", []string{"Describe", "Invite"}},
					{"ServiceAccount", []string{"Describe"}},
				},
			},
		},
	}

	environmentAdminRole := mdsv2alpha1.Role{
		"EnvironmentAdmin",
		[]mdsv2alpha1.AccessPolicy{
			{
				"ENVIRONMENT",
				false,
				[]mdsv2alpha1.Operation{
					{"SecurityMetadata", []string{"Describe", "Alter"}},
					{"ClusterApiKey", []string{"All"}},
					{"Connector", []string{"All"}},
					{"NetworkAccess", []string{"All"}},
					{"KsqlCluster", []string{"All"}},
					{"Environment", []string{"Alter", "Delete", "AlterAccess", "CreateKafkaCluster", "DescribeAccess"}},
					{"Subject", []string{"All"}},
					{"NetworkConfig", []string{"All"}},
					{"ClusterMetric", []string{"All"}},
					{"Cluster", []string{"All"}},
					{"SchemaRegistry", []string{"All"}},
					{"NetworkRegion", []string{"All"}},
					{"Deployment", []string{"All"}},
					{"Topic", []string{"All"}},
				},
			},
			{
				"organization",
				false,
				[]mdsv2alpha1.Operation{
					{"User", []string{"Describe", "Invite"}},
					{"ServiceAccount", []string{"Describe"}},
					{"SupportPlan", []string{"Describe"}},
				},
			},
		},
	}

	organizationAdminRole := mdsv2alpha1.Role{
		"OrganizationAdmin",
		[]mdsv2alpha1.AccessPolicy{
			{
				"organization",
				false,
				[]mdsv2alpha1.Operation{
					{"Topic", []string{"All"}},
					{"NetworkConfig", []string{"All"}},
					{"SecurityMetadata", []string{"Describe", "Alter"}},
					{"Billing", []string{"All"}},
					{"ClusterApiKey", []string{"All"}},
					{"Deployment", []string{"All"}},
					{"SchemaRegistry", []string{"All"}},
					{"KsqlCluster", []string{"All"}},
					{"CloudApiKey", []string{"All"}},
					{"NetworkAccess", []string{"All"}},
					{"SecuritySSO", []string{"All"}},
					{"SupportPlan", []string{"All"}},
					{"Connector", []string{"All"}},
					{"ClusterMetric", []string{"All"}},
					{"ServiceAccount", []string{"All"}},
					{"Subject", []string{"All"}},
					{"Cluster", []string{"All"}},
					{"Environment", []string{"All"}},
					{"NetworkRegion", []string{"All"}},
					{"Organization", []string{"Alter", "CreateEnvironment", "AlterAccess", "DescribeAccess"}},
					{"User", []string{"All"}},
				},
			},
		},
	}

	resourceOwnerRole := mdsv2alpha1.Role{
		"ResourceOwner",
		[]mdsv2alpha1.AccessPolicy{
			{
				"cloud-cluster",
				false,
				[]mdsv2alpha1.Operation{
					{"CloudCluster", []string{"Describe"}},
				},
			},
			{
				"cluster",
				true,
				[]mdsv2alpha1.Operation{
					{"Topic", []string{"Create", "Delete", "Read", "Write", "Describe", "DescribeConfigs", "Alter", "AlterConfigs", "DescribeAccess", "AlterAccess"}},
					{"Group", []string{"Read", "Describe", "Delete", "DescribeAccess", "AlterAccess"}},
				},
			},
		},
	}

	return []mdsv2alpha1.Role{ccloudRoleBindingAdminRole, cloudClusterAdminRole, environmentAdminRole, organizationAdminRole, resourceOwnerRole}
}

func rolesListToJsonMap(roles []mdsv2alpha1.Role) map[string]string {
	roleMap := make(map[string]string)
	for _, role := range roles {
		jsonVal, _ := json.Marshal(role)
		roleMap[role.Name] = string(jsonVal)
	}

	return roleMap
}
