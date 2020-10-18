package iam

import (
	"context"
	"fmt"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"net/http"
	"sort"
	"strings"

	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	resourcePatternListFields           = []string{"Principal", "Role", "ResourceType", "Name", "PatternType"}
	resourcePatternHumanListLabels      = []string{"Principal", "Role", "ResourceType", "Name", "PatternType"}
	resourcePatternStructuredListLabels = []string{"principal", "role", "resource_type", "name", "pattern_type"}

	// ccloud has Email as additional field
	ccloudResourcePatternListFields           = []string{"Principal", "Email", "Role", "ResourceType", "Name", "PatternType"}
	ccloudResourcePatternHumanListLabels      = []string{"Principal", "Email", "Role", "ResourceType", "Name", "PatternType"}
	ccloudResourcePatternStructuredListLabels = []string{"principal", "email", "role", "resource_type", "name", "pattern_type"}

	//TODO: please move this to a backend route
	clusterScopedRoles = map[string]bool{
		"SystemAdmin":   true,
		"ClusterAdmin":  true,
		"SecurityAdmin": true,
		"UserAdmin":     true,
		"Operator":      true,
	}

	clusterScopedRolesV2 = map[string]bool{
		"CloudClusterAdmin": true,
	}

	environmentScopedRoles = map[string]bool{
		"EnvironmentAdmin": true,
	}

	organizationScopedRoles = map[string]bool{
		"OrganizationAdmin": true,
	}
)

type rolebindingOptions struct {
	role             string
	resource         string
	prefix           bool
	principal        string
	scopeV2          mdsv2alpha1.Scope
	mdsScope         mds.MdsScope
	resourcesRequest mds.ResourcesRequest
}

type rolebindingCommand struct {
	*cmd.AuthenticatedCLICommand
	cliName string
}

type listDisplay struct {
	Principal    string
	Email        string
	Role         string
	ResourceType string
	Name         string
	PatternType  string
}

// NewRolebindingCommand returns the sub-command object for interacting with RBAC rolebindings.
func NewRolebindingCommand(cliName string, prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "rolebinding",
			Short: "Manage RBAC and IAM role bindings.",
			Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) role bindings.",
		}, prerunner)
	roleBindingCmd := &rolebindingCommand{
		AuthenticatedCLICommand: cliCmd,
		cliName:                 cliName,
	}
	roleBindingCmd.init()
	return roleBindingCmd.Command
}

func (c *rolebindingCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List role bindings.",
		Long:  "List the role bindings for a particular principal and/or role, and a particular scope.",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Only use the ``--resource`` flag when specifying a ``--role`` with no ``--principal`` specified. If specifying a ``--principal``, then the ``--resource`` flag is ignored. To list role bindings for a specific role on an identified resource:",
				Code: "iam rolebinding list --kafka-cluster-id CID  --role DeveloperRead --resource Topic",
			},
			examples.Example{
				Text: "To list the role bindings for a specific principal:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --principal User:frodo",
			},
			examples.Example{
				Text: "To list the role bindings for a specific principal, filtered to a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --principal User:frodo --role DeveloperRead",
			},
			examples.Example{
				Text: "To list the principals bound to a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --role DeveloperWrite",
			},
			examples.Example{
				Text: "To list the principals bound to a specific resource with a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --role DeveloperWrite --resource Topic:shire-parties",
			},
		),
	}
	listCmd.Flags().String("principal", "", "Principal whose rolebindings should be listed.")
	listCmd.Flags().Bool("current-user", false, "Show rolebindings belonging to current user.")
	listCmd.Flags().String("role", "", "List rolebindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.")
	if c.cliName == "ccloud" {
		listCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
		listCmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: "To list the role bindings for current user:",
				Code: "iam rolebinding list --current-user",
			},
			examples.Example{
				Text: "To list the role bindings for a specific principal:",
				Code: "iam rolebinding list --principal User:frodo",
			},
			examples.Example{
				Text: "To list the role bindings for a specific principal, filtered to a specific role:",
				Code: "iam rolebinding list --principal User:frodo --role CloudClusterAdmin --environment current --cloud-cluster lkc-1111aaa",
			},
			examples.Example{
				Text: "To list the principals bound to a specific role",
				Code: "iam rolebinding list --role CloudClusterAdmin --environment current --cloud-cluster lkc-1111aaa",
			},
		)
	} else {
		listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("resource", "", "If specified with a role and no principals, list principals with rolebindings to the role for this qualified resource.")
		listCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role binding.",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.create),
	}
	createCmd.Flags().String("role", "", "Role name of the new role binding.")
	createCmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	if c.cliName == "ccloud" {
		createCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		createCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
	} else {
		createCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
		createCmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
		createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		createCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		createCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
		createCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		createCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false
	check(createCmd.MarkFlagRequired("role"))
	check(createCmd.MarkFlagRequired("principal"))
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing role binding.",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.delete),
	}
	deleteCmd.Flags().String("role", "", "Role name of the existing role binding.")
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	if c.cliName == "ccloud" {
		deleteCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		deleteCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
	} else {
		deleteCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
		deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the role binding.")
		deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		deleteCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
		deleteCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		deleteCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	deleteCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	deleteCmd.Flags().SortFlags = false
	check(createCmd.MarkFlagRequired("role"))
	check(deleteCmd.MarkFlagRequired("principal"))
	c.AddCommand(deleteCmd)
}

func (c *rolebindingCommand) validatePrincipalFormat(principal string) error {
	if len(strings.Split(principal, ":")) == 1 {
		return errors.NewErrorWithSuggestions(errors.PrincipalFormatErrorMsg, errors.PrincipalFormatSuggestions)
	}

	return nil
}

func (c *rolebindingCommand) parseAndValidateResourcePattern(typename string, prefix bool) (mds.ResourcePattern, error) {
	var result mds.ResourcePattern
	if prefix {
		result.PatternType = "PREFIXED"
	} else {
		result.PatternType = "LITERAL"
	}

	parts := strings.Split(typename, ":")
	if len(parts) != 2 {
		return result, errors.NewErrorWithSuggestions(errors.ResourceFormatErrorMsg, errors.ResourceFormatSuggestions)
	}
	result.ResourceType = parts[0]
	result.Name = parts[1]

	return result, nil
}

func (c *rolebindingCommand) validateRoleAndResourceType(roleName string, resourceType string) error {
	ctx := c.createContext()
	role, resp, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName)
	if err != nil || resp.StatusCode == 204 {
		if err == nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
		} else {
			return errors.NewWrapErrorWithSuggestions(err, fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
		}
	}

	var allResourceTypes []string
	found := false
	for _, operation := range role.AccessPolicy.AllowedOperations {
		allResourceTypes = append(allResourceTypes, operation.ResourceType)
		if operation.ResourceType == resourceType {
			found = true
			break
		}
	}

	if !found {
		suggestionsMsg := fmt.Sprintf(errors.InvalidResourceTypeSuggestions, strings.Join(allResourceTypes, ", "))
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.InvalidResourceTypeErrorMsg, resourceType), suggestionsMsg)
	}

	return nil
}

func (c *rolebindingCommand) parseAndValidateScope(cmd *cobra.Command) (*mds.MdsScope, error) {
	scope := &mds.MdsScopeClusters{}
	nonKafkaScopesSet := 0

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return nil, err
	}

	cmd.Flags().Visit(func(flag *pflag.Flag) {
		switch flag.Name {
		case "kafka-cluster-id":
			scope.KafkaCluster = flag.Value.String()
		case "schema-registry-cluster-id":
			scope.SchemaRegistryCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "ksql-cluster-id":
			scope.KsqlCluster = flag.Value.String()
			nonKafkaScopesSet++
		case "connect-cluster-id":
			scope.ConnectCluster = flag.Value.String()
			nonKafkaScopesSet++
		}
	})

	if clusterName != "" && (scope.KafkaCluster != "" || nonKafkaScopesSet > 0) {
		return nil, errors.New(errors.BothClusterNameAndScopeErrorMsg)
	}

	if clusterName == "" {
		if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
			return nil, errors.New(errors.SpecifyKafkaIDErrorMsg)
		}

		if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
			return nil, errors.New(errors.SpecifyClusterErrorMsg)
		}

		if nonKafkaScopesSet > 1 {
			return nil, errors.New(errors.MoreThanOneNonKafkaErrorMsg)
		}
		return &mds.MdsScope{Clusters: *scope}, nil
	}

	return &mds.MdsScope{ClusterName: clusterName}, nil
}

func (c *rolebindingCommand) parseAndValidateScopeV2(cmd *cobra.Command) (*mdsv2alpha1.Scope, error) {
	scopeV2 := &mdsv2alpha1.Scope{}
	orgResourceId := c.State.Auth.Organization.GetResourceId()
	scopeV2.Path = []string{"organization=" + orgResourceId}

	if cmd.Flags().Changed("environment") {
		env, err := cmd.Flags().GetString("environment")
		if err != nil {
			return nil, err
		}
		if env == "current" {
			scopeV2.Path = append(scopeV2.Path, "environment="+c.EnvironmentId())
		} else {
			scopeV2.Path = append(scopeV2.Path, "environment="+env)
		}
	}

	if cmd.Flags().Changed("cloud-cluster") {
		cluster, err := cmd.Flags().GetString("cloud-cluster")
		if err != nil {
			return nil, err
		}
		scopeV2.Path = append(scopeV2.Path, "cloud-cluster="+cluster)
	}

	if cmd.Flags().Changed("role") {
		role, err := cmd.Flags().GetString("role")
		if err != nil {
			return nil, err
		}
		if clusterScopedRolesV2[role] && !cmd.Flags().Changed("cloud-cluster") {
			return nil, errors.New(errors.SpecifyCloudClusterErrorMsg)
		}
		if (environmentScopedRoles[role] || clusterScopedRolesV2[role]) && !cmd.Flags().Changed("environment") {
			return nil, errors.New(errors.SpecifyEnvironmentErrorMsg)
		}
	}

	if cmd.Flags().Changed("cloud-cluster") && !cmd.Flags().Changed("environment") {
		return nil, errors.New(errors.SpecifyCloudClusterErrorMsg)
	}
	return scopeV2, nil
}

func (c *rolebindingCommand) confluentList(cmd *cobra.Command, options *rolebindingOptions) error {
	if cmd.Flags().Changed("principal") {
		return c.listPrincipalResources(cmd, options)
	} else if cmd.Flags().Changed("role") {
		return c.confluentListRolePrincipals(cmd, options)
	}
	return errors.New(errors.PrincipalOrRoleRequiredErrorMsg)
}

func (c *rolebindingCommand) listMyRoleBindings(cmd *cobra.Command, options *rolebindingOptions) error {
	scopeV2 := &options.scopeV2
	var principal string
	currentUser, err := cmd.Flags().GetBool("current-user")
	if err != nil {
		return err
	}
	if currentUser {
		principal = "User:" + c.State.Auth.User.ResourceId
	} else {
		principal = options.principal
	}
	scopedRoleBindingMappings, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.MyRoleBindings(
		c.createContext(),
		principal,
		*scopeV2)
	if err != nil {
		return err
	}

	userToEmailMap, err := c.userIdToEmailMap()
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, ccloudResourcePatternListFields, ccloudResourcePatternHumanListLabels, ccloudResourcePatternStructuredListLabels)
	if err != nil {
		return err
	}

	for _, scopedRoleBindingMapping := range scopedRoleBindingMappings {
		roleBindingScope := scopedRoleBindingMapping.Scope
		for principalName, roleBindings := range scopedRoleBindingMapping.Rolebindings {
			principalEmail := userToEmailMap[principalName]
			for roleName, resourcePatterns := range roleBindings {
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
						Principal:    principalName,
						Email:        principalEmail,
						Role:         roleName,
						ResourceType: resourcePattern.ResourceType,
						Name:         resourcePattern.Name,
						PatternType:  resourcePattern.PatternType,
					})
				}
				if cmd.Flags().Changed("role") {
					role, err := cmd.Flags().GetString("role")
					if err != nil {
						return err
					}
					if role != roleName {
						continue
					}
				}
				orgName := ""
				envName := ""
				clusterName := ""
				for _, elem := range roleBindingScope.Path {
					if strings.HasPrefix(elem, "organization=") {
						orgName = strings.TrimPrefix(elem, "organization=")
					}
					if strings.HasPrefix(elem, "environment=") {
						envName = strings.TrimPrefix(elem, "environment=")
					}
					if strings.HasPrefix(elem, "cloud-cluster=") {
						clusterName = strings.TrimPrefix(elem, "cloud-cluster=")
					}
				}
				if len(resourcePatterns) == 0 && organizationScopedRoles[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Email:        principalEmail,
						Role:         roleName,
						ResourceType: "Organization",
						Name:         orgName,
						PatternType:  "",
					})
				}
				if len(resourcePatterns) == 0 && environmentScopedRoles[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Email:        principalEmail,
						Role:         roleName,
						ResourceType: "Environment",
						Name:         envName,
						PatternType:  "",
					})
				}
				if len(resourcePatterns) == 0 && clusterScopedRolesV2[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Email:        principalEmail,
						Role:         roleName,
						ResourceType: "Cluster",
						Name:         clusterName,
						PatternType:  "",
					})
				}
			}
		}
	}

	outputWriter.StableSort()

	return outputWriter.Out()
}

func (c *rolebindingCommand) ccloudList(cmd *cobra.Command, options *rolebindingOptions) error {
	if cmd.Flags().Changed("principal") || cmd.Flags().Changed("current-user") {
		return c.listMyRoleBindings(cmd, options)
	} else if cmd.Flags().Changed("role") {
		return c.ccloudListRolePrincipals(cmd, options)
	} else {
		return errors.New(errors.PrincipalOrRoleRequiredErrorMsg)
	}
}

func (c *rolebindingCommand) list(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}
	if c.cliName == "ccloud" {
		return c.ccloudList(cmd, options)
	} else {
		return c.confluentList(cmd, options)
	}
}

func (c *rolebindingCommand) listPrincipalResources(cmd *cobra.Command, options *rolebindingOptions) error {
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
	principalsRolesResourcePatterns, response, err := c.MDSClient.RBACRoleBindingSummariesApi.LookupResourcesForPrincipal(
		c.createContext(),
		principal,
		*scope)
	if err != nil {
		if response.StatusCode == http.StatusNotFound {
			return c.listPrincipalResourcesV1(cmd, scope, principal, role)
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
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Role:         roleName,
						ResourceType: resourcePattern.ResourceType,
						Name:         resourcePattern.Name,
						PatternType:  resourcePattern.PatternType,
					})
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

func (c *rolebindingCommand) listPrincipalResourcesV1(cmd *cobra.Command, mdsScope *mds.MdsScope, principal string, role string) error {
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

func (c *rolebindingCommand) confluentListRolePrincipals(cmd *cobra.Command, options *rolebindingOptions) error {
	scope := &options.mdsScope
	role := options.role

	var principals []string
	var err error
	if cmd.Flags().Changed("resource") {
		r, err := cmd.Flags().GetString("resource")
		if err != nil {
			return err
		}
		resource, err := c.parseAndValidateResourcePattern(r, false)
		if err != nil {
			return err
		}
		err = c.validateRoleAndResourceType(role, resource.ResourceType)
		if err != nil {
			return err
		}
		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRoleOnResource(
			c.createContext(),
			role,
			resource.ResourceType,
			resource.Name,
			*scope)
		if err != nil {
			return err
		}
	} else {
		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(
			c.createContext(),
			role,
			*scope)
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

func (c *rolebindingCommand) ccloudListRolePrincipals(cmd *cobra.Command, options *rolebindingOptions) error {
	scopeV2 := &options.scopeV2
	role := options.role

	principals, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(
		c.createContext(),
		role,
		*scopeV2)
	if err != nil {
		return err
	}

	userToEmailMap, err := c.userIdToEmailMap()
	if err != nil {
		return err
	}

	sort.Strings(principals)
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Principal", "Email"}, []string{"Principal", "Email"}, []string{"principal", "email"})
	if err != nil {
		return err
	}
	for _, principal := range principals {
		displayStruct := &struct {
			Principal string
			Email     string
		}{
			Principal: principal,
			Email:     userToEmailMap[principal],
		}
		outputWriter.AddElement(displayStruct)
	}
	return outputWriter.Out()
}

func (c *rolebindingCommand) userIdToEmailMap() (map[string]string, error) {
	userToEmailMap := make(map[string]string)
	users, err := c.Client.User.List(context.Background())
	if err != nil {
		return userToEmailMap, err
	}
	for _, u := range users {
		userToEmailMap["User:"+u.ResourceId] = u.Email
	}
	return userToEmailMap, nil
}

func (c *rolebindingCommand) parseCommon(cmd *cobra.Command) (*rolebindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	resource := ""
	prefix := false
	if c.cliName != "ccloud" {
		resource, err = cmd.Flags().GetString("resource")
		if err != nil {
			return nil, err
		}
		prefix = cmd.Flags().Changed("prefix")
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(principal, "User:") {
		principalValue := strings.TrimLeft(principal, "User:")
		if strings.Contains(principalValue, "@") {
			user, err := c.Client.User.Describe(context.Background(), &orgv1.User{Email: principalValue, OrganizationId: c.State.Auth.Organization.GetId()})
			if err != nil {
				return nil, err
			}
			principal = "User:" + user.ResourceId
		}
	}
	if cmd.Flags().Changed("principal") {
		err = c.validatePrincipalFormat(principal)
		if err != nil {
			return nil, err
		}
	}

	scope := &mds.MdsScope{}
	scopeV2 := &mdsv2alpha1.Scope{}
	if c.cliName != "ccloud" {
		scope, err = c.parseAndValidateScope(cmd)
	} else {
		scopeV2, err = c.parseAndValidateScopeV2(cmd)
	}
	if err != nil {
		return nil, err
	}

	resourcesRequest := mds.ResourcesRequest{}
	if resource != "" {
		parsedResourcePattern, err := c.parseAndValidateResourcePattern(resource, prefix)
		if err != nil {
			return nil, err
		}
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		if err != nil {
			return nil, err
		}
		resourcePatterns := []mds.ResourcePattern{
			parsedResourcePattern,
		}
		resourcesRequest = mds.ResourcesRequest{
			Scope:            *scope,
			ResourcePatterns: resourcePatterns,
		}
	}
	return &rolebindingOptions{
			role,
			resource,
			prefix,
			principal,
			*scopeV2,
			*scope,
			resourcesRequest,
		},
		nil
}

func (c *rolebindingCommand) confluentCreate(options *rolebindingOptions) (resp *http.Response, err error) {
	if options.resource != "" {
		resp, err = c.MDSClient.RBACRoleBindingCRUDApi.AddRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else {
		resp, err = c.MDSClient.RBACRoleBindingCRUDApi.AddRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.mdsScope)
	}
	return
}

func (c *rolebindingCommand) ccloudCreate(options *rolebindingOptions) (*http.Response, error) {
	return c.MDSv2Client.RBACRoleBindingCRUDApi.AddRoleForPrincipal(
		c.createContext(),
		options.principal,
		options.role,
		options.scopeV2)
}

func (c *rolebindingCommand) create(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	var resp *http.Response
	if c.cliName == "ccloud" {
		resp, err = c.ccloudCreate(options)
	} else {
		resp, err = c.confluentCreate(options)
	}

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, resp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}
	if c.cliName == "ccloud" {
		return c.displayCCloudCreateAndDeleteOutput(cmd, options)
	} else {
		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *rolebindingCommand) displayCCloudCreateAndDeleteOutput(cmd *cobra.Command, options *rolebindingOptions) error {
	var fieldsSelected []string
	structuredRename := map[string]string{"Principal": "principal", "Email": "email", "Role": "role"}
	userResourceId := strings.TrimLeft(options.principal, "User:")
	user, err := c.Client.User.Describe(context.Background(), &orgv1.User{ResourceId: userResourceId, OrganizationId: c.State.Auth.Organization.GetId()})
	displayStruct := &listDisplay{
		Principal: options.principal,
		Role:      options.role,
	}
	if err != nil {
		fieldsSelected = []string{"Principal", "Role"}
	} else {
		displayStruct.Email = user.Email
		fieldsSelected = []string{"Principal", "Email", "Role"}
	}
	return output.DescribeObject(cmd, displayStruct, fieldsSelected, map[string]string{}, structuredRename)
}

func displayCreateAndDeleteOutput(cmd *cobra.Command, options *rolebindingOptions) error {
	var fieldsSelected []string
	structuredRename := map[string]string{"Principal": "principal", "Role": "role", "ResourceType": "resource_type", "Name": "name", "PatternType": "pattern_type"}
	displayStruct := &listDisplay{
		Principal: options.principal,
		Role:      options.role,
	}
	if options.resource != "" {
		fieldsSelected = resourcePatternListFields
		if len(options.resourcesRequest.ResourcePatterns) != 1 {
			return errors.New("display error: number of resource pattern is not 1")
		}
		resourcePattern := options.resourcesRequest.ResourcePatterns[0]
		displayStruct.ResourceType = resourcePattern.ResourceType
		displayStruct.Name = resourcePattern.Name
		displayStruct.PatternType = resourcePattern.PatternType
	} else {
		fieldsSelected = []string{"Principal", "Role", "ResourceType"}
		displayStruct.ResourceType = "Cluster"
	}
	return output.DescribeObject(cmd, displayStruct, fieldsSelected, map[string]string{}, structuredRename)
}

func (c *rolebindingCommand) confluentDelete(options *rolebindingOptions) (resp *http.Response, err error) {
	if options.resource != "" {
		resp, err = c.MDSClient.RBACRoleBindingCRUDApi.RemoveRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else {
		resp, err = c.MDSClient.RBACRoleBindingCRUDApi.DeleteRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.mdsScope)
	}
	return
}

func (c *rolebindingCommand) ccloudDelete(options *rolebindingOptions) (*http.Response, error) {
	return c.MDSv2Client.RBACRoleBindingCRUDApi.DeleteRoleForPrincipal(
		c.createContext(),
		options.principal,
		options.role,
		options.scopeV2)
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	var resp *http.Response
	if c.cliName == "ccloud" {
		resp, err = c.ccloudDelete(options)
	} else {
		resp, err = c.confluentDelete(options)
	}

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, resp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	if c.cliName == "ccloud" {
		return c.displayCCloudCreateAndDeleteOutput(cmd, options)
	} else {
		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (c *rolebindingCommand) createContext() context.Context {
	if c.cliName == "ccloud" {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.AuthToken())
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.AuthToken())
	}
}
