package iam

import (
	"context"
	"net/http"
	"os"
	"sort"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"

	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type displayStruct struct {
	Principal   string
	Email       string
	ServiceName string
	PoolName    string
}

func (c *roleBindingCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List role bindings.",
		Long:  "List the role bindings for a particular principal and/or role, and a particular scope.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	if c.cfg.IsCloudLogin() {
		cmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: "List the role bindings for current user:",
				Code: "confluent iam rbac role-binding list --current-user",
			},
			examples.Example{
				Text: `List the role bindings for user "u-123456":`,
				Code: "confluent iam rbac role-binding list --principal User:u-123456",
			},
			examples.Example{
				Text: `List the role bindings for principals with role "CloudClusterAdmin":`,
				Code: "confluent iam rbac role-binding list --role CloudClusterAdmin --current-env --cloud-cluster lkc-123456",
			},
			examples.Example{
				Text: `List the role bindings for user "u-123456" with role "CloudClusterAdmin":`,
				Code: "confluent iam rbac role-binding list --principal User:u-123456 --role CloudClusterAdmin --environment env-12345 --cloud-cluster lkc-123456",
			},
		)
	} else {
		cmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: "Only use the `--resource` flag when specifying a `--role` with no `--principal` specified. If specifying a `--principal`, then the `--resource` flag is ignored. To list role bindings for a specific role on an identified resource:",
				Code: "confluent iam rbac role-binding list --kafka-cluster-id $KAFKA_CLUSTER_ID --role DeveloperRead --resource Topic",
			},
			examples.Example{
				Text: "List the role bindings for a specific principal:",
				Code: "confluent iam rbac role-binding list --kafka-cluster-id $KAFKA_CLUSTER_ID --principal User:my-user",
			},
			examples.Example{
				Text: "List the role bindings for a specific principal, filtered to a specific role:",
				Code: "confluent iam rbac role-binding list --kafka-cluster-id $KAFKA_CLUSTER_ID --principal User:my-user --role DeveloperRead",
			},
			examples.Example{
				Text: "List the principals bound to a specific role:",
				Code: "confluent iam rbac role-binding list --kafka-cluster-id $KAFKA_CLUSTER_ID --role DeveloperWrite",
			},
			examples.Example{
				Text: "List the principals bound to a specific resource with a specific role:",
				Code: "confluent iam rbac role-binding list --kafka-cluster-id $KAFKA_CLUSTER_ID --role DeveloperWrite --resource Topic:my-topic",
			},
		)
	}

	cmd.Flags().String("principal", "", "Principal whose role bindings should be listed.")
	cmd.Flags().Bool("current-user", false, "Show role bindings belonging to current user.")
	cmd.Flags().String("role", "", "List role bindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.")

	if c.cfg.IsCloudLogin() {
		cmd.Flags().String("environment", "", "Environment ID for scope of role binding listings.")
		cmd.Flags().Bool("current-env", false, "Use current environment ID for scope.")
		cmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for scope of role binding listings.")
		cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of role binding listings.")
		if os.Getenv("XX_DATAPLANE_3_ENABLE") != "" {
			cmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding listings.")
			cmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding listings.")
		}
	} else {
		cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of role binding listings.")
		cmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of role binding listings.")
		cmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for scope of role binding listings.")
		cmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of role binding listings.")
		cmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for role binding listings.")
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}

	cmd.Flags().String("resource", "", "If specified with a role and no principals, list principals with role bindings to the role for this qualified resource.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *roleBindingCommand) list(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	if c.cfg.IsCloudLogin() {
		return c.ccloudList(cmd, options)
	} else {
		return c.confluentList(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudList(cmd *cobra.Command, options *roleBindingOptions) error {
	if cmd.Flags().Changed("principal") || cmd.Flags().Changed("current-user") {
		return c.listMyRoleBindings(cmd, options)
	} else if cmd.Flags().Changed("role") {
		return c.ccloudListRolePrincipals(cmd, options)
	}
	return errors.New(errors.PrincipalOrRoleRequiredErrorMsg)
}

func (c *roleBindingCommand) listMyRoleBindings(cmd *cobra.Command, options *roleBindingOptions) error {
	scopeV2 := &options.scopeV2

	currentUser, err := cmd.Flags().GetBool("current-user")
	if err != nil {
		return err
	}

	principal := options.principal
	if currentUser {
		principal = "User:" + c.Context.GetUser().GetResourceId()
	}

	scopedRoleBindingMappings, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.MyRoleBindings(c.createContext(), principal, *scopeV2)
	if err != nil {
		return err
	}

	userToEmailMap, err := c.getUserIdToEmailMap()
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, ccloudResourcePatternListFields, ccloudResourcePatternHumanListLabels, ccloudResourcePatternStructuredListLabels)
	if err != nil {
		return err
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return err
	}

	for _, scopedRoleBindingMapping := range scopedRoleBindingMappings {
		roleBindingScope := scopedRoleBindingMapping.Scope
		for principalName, roleBindings := range scopedRoleBindingMapping.Rolebindings {
			principalEmail := userToEmailMap[principalName]
			for roleName, resourcePatterns := range roleBindings {
				if role != "" && role != roleName {
					continue
				}

				envName := ""
				cloudClusterName := ""
				for _, elem := range roleBindingScope.Path {
					// we don't capture the organization name because it's always this organization
					if strings.HasPrefix(elem, "environment=") {
						envName = strings.TrimPrefix(elem, "environment=")
					}
					if strings.HasPrefix(elem, "cloud-cluster=") {
						cloudClusterName = strings.TrimPrefix(elem, "cloud-cluster=")
					}
				}
				clusterType := ""
				logicalCluster := ""
				if roleBindingScope.Clusters.ConnectCluster != "" {
					clusterType = "Connect"
					logicalCluster = roleBindingScope.Clusters.ConnectCluster
				} else if roleBindingScope.Clusters.KsqlCluster != "" {
					clusterType = "ksqlDB"
					clusterName := roleBindingScope.Clusters.KsqlCluster
					req := &schedv1.KSQLCluster{AccountId: c.EnvironmentId()}
					clusterList, err := c.Client.KSQL.List(context.Background(), req)
					if err != nil {
						return err
					}
					for _, ksql := range clusterList {
						if ksql.KafkaClusterId == cloudClusterName && ksql.Name == clusterName {
							logicalCluster = ksql.Id
							break
						}
					}
					// When empty, fill it up as printout depends on this
					if len(resourcePatterns) == 0 {
						resourcePatterns = append(resourcePatterns, mdsv2alpha1.ResourcePattern{
							ResourceType: "KSQL",
							Name:         clusterName,
							PatternType:  "LITERAL",
						})
					}
				} else if roleBindingScope.Clusters.SchemaRegistryCluster != "" {
					clusterType = "Schema Registry"
					logicalCluster = roleBindingScope.Clusters.SchemaRegistryCluster
				} else if roleBindingScope.Clusters.KafkaCluster != "" {
					clusterType = "Kafka"
					logicalCluster = roleBindingScope.Clusters.KafkaCluster
				}

				for _, resourcePattern := range resourcePatterns {
					if cmd.Flags().Changed("resource") {
						resource, err := cmd.Flags().GetString("resource")
						if err != nil {
							return err
						}
						if resource != resourcePattern.ResourceType {
							continue
						}
					}
					outputWriter.AddElement(&listDisplay{
						Principal:      principalName,
						Email:          principalEmail,
						Role:           roleName,
						Environment:    envName,
						CloudCluster:   cloudClusterName,
						ClusterType:    clusterType,
						LogicalCluster: logicalCluster,
						ResourceType:   resourcePattern.ResourceType,
						Name:           resourcePattern.Name,
						PatternType:    resourcePattern.PatternType,
					})
				}

				if len(resourcePatterns) == 0 {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Email:        principalEmail,
						Role:         roleName,
						Environment:  envName,
						CloudCluster: cloudClusterName,
					})
				}
			}
		}
	}

	outputWriter.StableSort()

	return outputWriter.Out()
}

func (c *roleBindingCommand) getPoolToNameMap() (map[string]string, error) {
	providers, err := c.V2Client.ListIdentityProviders()
	if err != nil {
		return map[string]string{}, err
	}
	poolToName := make(map[string]string)
	for _, provider := range providers {
		pools, err := c.V2Client.ListIdentityPools(*provider.Id)
		if err != nil {
			return map[string]string{}, err
		}
		for _, pool := range pools {
			poolToName["User:"+*pool.Id] = *pool.DisplayName
		}
	}
	return poolToName, nil
}

func (c *roleBindingCommand) getUserIdToEmailMap() (map[string]string, error) {
	users, err := c.V2Client.ListIamUsers()
	if err != nil {
		return nil, err
	}
	userToEmail := make(map[string]string)
	for _, u := range users {
		userToEmail["User:"+u.GetId()] = u.GetEmail()
	}
	return userToEmail, nil
}

func (c *roleBindingCommand) getServiceAccountIdToNameMap() (map[string]string, error) {
	serviceAccounts, err := c.V2Client.ListIamServiceAccounts()
	if err != nil {
		return nil, err
	}
	serviceAccountToName := make(map[string]string)
	for _, sa := range serviceAccounts {
		serviceAccountToName["User:"+sa.GetId()] = sa.GetDisplayName()
	}
	return serviceAccountToName, nil
}

func (c *roleBindingCommand) ccloudListRolePrincipals(cmd *cobra.Command, options *roleBindingOptions) error {
	scopeV2 := &options.scopeV2
	role := options.role

	var principals []string
	var err error
	if cmd.Flags().Changed("resource") {
		r, err := cmd.Flags().GetString("resource")
		if err != nil {
			return err
		}
		resource, err := parseAndValidateResourcePatternV2(r, false)
		if err != nil {
			return err
		}
		err = c.validateRoleAndResourceTypeV2(role, resource.ResourceType)
		if err != nil {
			return err
		}
		principals, _, err = c.MDSv2Client.RBACRoleBindingSummariesApi.LookupPrincipalsWithRoleOnResource(
			c.createContext(),
			role,
			resource.ResourceType,
			resource.Name,
			*scopeV2)
		if err != nil {
			return err
		}
	} else {
		principals, _, err = c.MDSv2Client.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(
			c.createContext(),
			role,
			*scopeV2)
		if err != nil {
			return err
		}
	}

	userToEmailMap, err := c.getUserIdToEmailMap()
	if err != nil {
		return err
	}

	serviceAccountToNameMap, err := c.getServiceAccountIdToNameMap()
	if err != nil {
		return err
	}

	// TODO: Catch this error once Identity Providers goes GA
	poolToNameMap, _ := c.getPoolToNameMap()

	sort.Strings(principals)
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Principal", "Email", "ServiceName", "PoolName"}, []string{"Principal", "Email", "Service Name", "Pool Name"}, []string{"principal", "email", "service_name", "pool_name"})
	if err != nil {
		return err
	}
	for _, principal := range principals {
		row := &displayStruct{Principal: principal}
		if email, ok := userToEmailMap[principal]; ok {
			row.Email = email
			outputWriter.AddElement(row)
		}
		if name, ok := serviceAccountToNameMap[principal]; ok {
			row.ServiceName = name
			outputWriter.AddElement(row)
		}
		if name, ok := poolToNameMap[principal]; ok {
			row.PoolName = name
			outputWriter.AddElement(row)
		}
	}
	return outputWriter.Out()
}

func (c *roleBindingCommand) confluentList(cmd *cobra.Command, options *roleBindingOptions) error {
	if cmd.Flags().Changed("principal") {
		return c.listPrincipalResources(cmd, options)
	} else if cmd.Flags().Changed("role") {
		return c.confluentListRolePrincipals(cmd, options)
	}
	return errors.New(errors.PrincipalOrRoleRequiredErrorMsg)
}

func (c *roleBindingCommand) listPrincipalResources(cmd *cobra.Command, options *roleBindingOptions) error {
	scope := &options.mdsScope
	principal := options.principal

	role := "*"
	if cmd.Flags().Changed("role") {
		r, err := cmd.Flags().GetString("role")
		if err != nil {
			return err
		}
		role = r
	}

	principalsRolesResourcePatterns, response, err := c.MDSClient.RBACRoleBindingSummariesApi.LookupResourcesForPrincipal(c.createContext(), principal, *scope)
	if err != nil {
		if response != nil && response.StatusCode == http.StatusNotFound {
			return c.listPrincipalResourcesV1(scope, principal, role)
		}
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, resourcePatternListFields, resourcePatternHumanListLabels, resourcePatternStructuredListLabels)
	if err != nil {
		return err
	}

	for principalName, rolesResourcePatterns := range principalsRolesResourcePatterns {
		for roleName, resourcePatterns := range rolesResourcePatterns {
			if role == "*" || roleName == role {
				for _, resourcePattern := range resourcePatterns {
					add := true
					if options.resource != "" {
						add = false
						for _, rp := range options.resourcesRequest.ResourcePatterns {
							if rp == resourcePattern {
								add = true
							}
						}
					}
					if add {
						outputWriter.AddElement(&listDisplay{
							Principal:    principalName,
							Role:         roleName,
							ResourceType: resourcePattern.ResourceType,
							Name:         resourcePattern.Name,
							PatternType:  resourcePattern.PatternType,
						})
					}
				}
				if len(resourcePatterns) == 0 && clusterScopedRoles[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Role:         roleName,
						ResourceType: "Cluster",
						Name:         "",
						PatternType:  "",
					})
				}
			}
		}
	}

	outputWriter.StableSort()

	return outputWriter.Out()
}

func (c *roleBindingCommand) listPrincipalResourcesV1(mdsScope *mds.MdsScope, principal string, role string) error {
	var err error
	roleNames := []string{role}
	if role == "*" {
		roleNames, _, err = c.MDSClient.RBACRoleBindingSummariesApi.ScopedPrincipalRolenames(
			c.createContext(),
			principal,
			*mdsScope)
		if err != nil {
			return err
		}
	}

	var data [][]string
	for _, roleName := range roleNames {
		rps, _, err := c.MDSClient.RBACRoleBindingCRUDApi.GetRoleResourcesForPrincipal(
			c.createContext(),
			principal,
			roleName,
			*mdsScope)
		if err != nil {
			return err
		}
		for _, pattern := range rps {
			data = append(data, []string{roleName, pattern.ResourceType, pattern.Name, pattern.PatternType})
		}
		if len(rps) == 0 && clusterScopedRoles[roleName] {
			data = append(data, []string{roleName, "Cluster", "", ""})
		}
	}

	printer.RenderCollectionTable(data, []string{"Role", "ResourceType", "Name", "PatternType"})
	return nil
}

func (c *roleBindingCommand) confluentListRolePrincipals(cmd *cobra.Command, options *roleBindingOptions) error {
	scope := &options.mdsScope
	role := options.role

	var principals []string
	if cmd.Flags().Changed("resource") {
		r, err := cmd.Flags().GetString("resource")
		if err != nil {
			return err
		}

		resource, err := parseAndValidateResourcePattern(r, false)
		if err != nil {
			return err
		}

		if err := c.validateRoleAndResourceTypeV1(role, resource.ResourceType); err != nil {
			return err
		}

		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRoleOnResource(c.createContext(), role, resource.ResourceType, resource.Name, *scope)
		if err != nil {
			return err
		}
	} else {
		var err error
		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(c.createContext(), role, *scope)
		if err != nil {
			return err
		}
	}

	sort.Strings(principals)
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Principal"}, []string{"Principal"}, []string{"principal"})
	if err != nil {
		return err
	}

	for _, principal := range principals {
		displayStruct := &struct {
			Principal string
		}{
			Principal: principal,
		}
		outputWriter.AddElement(displayStruct)
	}
	return outputWriter.Out()
}
