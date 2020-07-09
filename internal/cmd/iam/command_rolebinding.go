package iam

import (
	"context"
	"fmt"
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
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) role bindings.",
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
		Args:  cobra.NoArgs,
		RunE:  c.list,
		Short: "List role bindings.",
		Long:  "List the role bindings for a particular principal and/or role, and a particular scope.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Only use the ``--resource`` flag when specifying a ``--role`` with no ``--principal`` specified. If specifying a ``--principal``, then the ``--resource`` flag is ignored. To list role bindings for a specific role on an identified resource:",
				Code: "iam rolebinding list --kafka-cluster-id CID  --role DeveloperRead --resource Topic",
			},
			examples.Example{
				Desc: "To list the role bindings for a specific principal:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --principal User:frodo",
			},
			examples.Example{
				Desc: "To list the role bindings for a specific principal, filtered to a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --principal User:frodo --role DeveloperRead",
			},
			examples.Example{
				Desc: "To list the principals bound to a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --role DeveloperWrite",
			},
			examples.Example{
				Desc: "To list the principals bound to a specific resource with a specific role:",
				Code: "iam rolebinding list --kafka-cluster-id $CID --role DeveloperWrite --resource Topic:shire-parties",
			},
		),
	}
	listCmd.Flags().String("principal", "", "Principal whose rolebindings should be listed.")
	listCmd.Flags().String("role", "", "List rolebindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.")
	if c.cliName == "ccloud" {
		listCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
	} else {
		listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("resource", "", "If specified with a role and no principals, list principals with rolebindings to the role for this qualified resource.")
		listCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of rolebinding listings.")
		listCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role binding.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("role", "", "Role name of the new role binding.")
	createCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	createCmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	if c.cliName == "ccloud" {
		createCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		createCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
	} else {
		createCmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
		createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		createCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		createCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
		createCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		createCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	createCmd.Flags().SortFlags = false
	check(createCmd.MarkFlagRequired("role"))
	check(createCmd.MarkFlagRequired("principal"))
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing role binding.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("role", "", "Role name of the existing role binding.")
	deleteCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	if c.cliName == "ccloud" {
		deleteCmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		deleteCmd.Flags().String("environment", "", "Environment ID for scope of rolebinding listings.")
	} else {
		deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the role binding.")
		deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		deleteCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
		deleteCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		deleteCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	}
	deleteCmd.Flags().SortFlags = false
	check(createCmd.MarkFlagRequired("role"))
	check(deleteCmd.MarkFlagRequired("principal"))
	c.AddCommand(deleteCmd)
}

func (c *rolebindingCommand) validatePrincipalFormat(principal string) error {
	if len(strings.Split(principal, ":")) == 1 {
		return errors.New("Principal must be specified in this format: <Principal Type>:<Principal Name>")
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
		return result, errors.New("Resource must be specified in this format: <Resource Type>:<Resource Name>")
	}
	result.ResourceType = parts[0]
	result.Name = parts[1]

	return result, nil
}

func (c *rolebindingCommand) parseAndValidateResourcePatternV2(typename string, prefix bool) (mdsv2alpha1.ResourcePattern, error) {
	r, err := c.parseAndValidateResourcePattern(typename, prefix)
	rv2 := mdsv2alpha1.ResourcePattern{
		PatternType: r.PatternType,
	}
	if err != nil {
		return rv2, err
	}
	rv2.Name = r.Name
	rv2.ResourceType = r.ResourceType
	return rv2, err
}

func (c *rolebindingCommand) validateRoleAndResourceType(roleName string, resourceType string) error {
	if c.cliName == "ccloud" {
		return nil
	}
	ctx := c.createContext()
	role, resp, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName)
	if err != nil || resp.StatusCode == 204 {
		return errors.Wrapf(err, "Failed to look up role %s. Was an invalid role name specified?", roleName)
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
		return errors.New("Invalid resource type " + resourceType + " specified. It must be one of " + strings.Join(allResourceTypes, ", "))
	}

	return nil
}

func (c *rolebindingCommand) parseAndValidateScope(cmd *cobra.Command) (*mds.MdsScope, error) {
	scope := &mds.MdsScopeClusters{}
	nonKafkaScopesSet := 0

	clusterName, err := cmd.Flags().GetString("cluster-name")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
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
		return nil, errors.HandleCommon(errors.New("Cannot specify both cluster name and cluster scope."), cmd)
	}

	if clusterName == "" {
		if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
			return nil, errors.HandleCommon(errors.New("Must also specify a --kafka-cluster-id to uniquely identify the scope."), cmd)
		}

		if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
			return nil, errors.HandleCommon(errors.New("Must specify either cluster ID flag to indicate role binding scope or the cluster name."), cmd)
		}

		if nonKafkaScopesSet > 1 {
			return nil, errors.HandleCommon(errors.New("Cannot specify more than one non-Kafka cluster ID for a scope."), cmd)
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
			return nil, errors.HandleCommon(err, cmd)
		}
		if clusterScopedRolesV2[role] && !cmd.Flags().Changed("cloud-cluster") {
			return nil, errors.HandleCommon(errors.New("Must specify cloud-cluster flag to indicate role binding scope."), cmd)
		}
		if (environmentScopedRoles[role] || clusterScopedRolesV2[role]) && !cmd.Flags().Changed("environment") {
			return nil, errors.HandleCommon(errors.New("Must specify environment flag to indicate role binding scope"), cmd)
		}
	}

	if cmd.Flags().Changed("cloud-cluster") && !cmd.Flags().Changed("environment") {
		return nil, errors.HandleCommon(errors.New("Must also specify environment flag to indicate role binding scope"), cmd)
	}
	return scopeV2, nil
}

func (c *rolebindingCommand) confluentList(cmd *cobra.Command) error {
	if cmd.Flags().Changed("principal") {
		return c.listPrincipalResources(cmd)
	} else if cmd.Flags().Changed("role") {
		return c.confluentListRolePrincipals(cmd)
	}
	return errors.HandleCommon(fmt.Errorf("required: either principal or role is required"), cmd)
}

func (c *rolebindingCommand) listMyRoleBindings(cmd *cobra.Command) error {
	scopeV2, err := c.parseAndValidateScopeV2(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if principal == "current" {
		principal = "User:" + c.State.Auth.User.ResourceId
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	scopedRoleBindingMappings, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.MyRoleBindings(
		c.createContext(),
		principal,
		*scopeV2)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, resourcePatternListFields, resourcePatternHumanListLabels, resourcePatternStructuredListLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	for _, scopedRoleBindingMapping := range scopedRoleBindingMappings {
		roleBindingScope := scopedRoleBindingMapping.Scope
		for principalName, roleBindings := range scopedRoleBindingMapping.Rolebindings {
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
						Role:         roleName,
						ResourceType: "Organization",
						Name:         orgName,
						PatternType:  "",
					})
				}
				if len(resourcePatterns) == 0 && environmentScopedRoles[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
						Role:         roleName,
						ResourceType: "Environment",
						Name:         envName,
						PatternType:  "",
					})
				}
				if len(resourcePatterns) == 0 && clusterScopedRolesV2[roleName] {
					outputWriter.AddElement(&listDisplay{
						Principal:    principalName,
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

func (c *rolebindingCommand) ccloudList(cmd *cobra.Command) error {
	if cmd.Flags().Changed("principal") {
		return c.listMyRoleBindings(cmd)
	} else if cmd.Flags().Changed("role") {
		return c.ccloudListRolePrincipals(cmd)
	} else {
		return errors.HandleCommon(errors.New("required: either principal or role is required"), cmd)
	}
}

func (c *rolebindingCommand) list(cmd *cobra.Command, _ []string) error {
	if c.cliName == "ccloud" {
		return c.ccloudList(cmd)
	} else {
		return c.confluentList(cmd)
	}
}

func (c *rolebindingCommand) listPrincipalResources(cmd *cobra.Command) error {
	scope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	role := "*"
	if cmd.Flags().Changed("role") {
		r, err := cmd.Flags().GetString("role")
		if err != nil {
			return errors.HandleCommon(err, cmd)
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
		return errors.HandleCommon(err, cmd)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, resourcePatternListFields, resourcePatternHumanListLabels, resourcePatternStructuredListLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
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
			return errors.HandleCommon(err, cmd)
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
			return errors.HandleCommon(err, cmd)
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

func (c *rolebindingCommand) confluentListRolePrincipals(cmd *cobra.Command) error {
	scope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var principals []string
	if cmd.Flags().Changed("resource") {
		r, err := cmd.Flags().GetString("resource")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		resource, err := c.parseAndValidateResourcePattern(r, false)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		err = c.validateRoleAndResourceType(role, resource.ResourceType)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRoleOnResource(
			c.createContext(),
			role,
			resource.ResourceType,
			resource.Name,
			*scope)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	} else {
		principals, _, err = c.MDSClient.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(
			c.createContext(),
			role,
			*scope)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	sort.Strings(principals)
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Principal"}, []string{"Principal"}, []string{"principal"})
	if err != nil {
		return errors.HandleCommon(err, cmd)
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

func (c *rolebindingCommand) ccloudListRolePrincipals(cmd *cobra.Command) error {
	scopeV2, err := c.parseAndValidateScopeV2(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	principals, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.LookupPrincipalsWithRole(
		c.createContext(),
		role,
		*scopeV2)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	sort.Strings(principals)
	outputWriter, err := output.NewListOutputWriter(cmd, []string{"Principal"}, []string{"Principal"}, []string{"principal"})
	if err != nil {
		return errors.HandleCommon(err, cmd)
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

func (c *rolebindingCommand) parseCommon(cmd *cobra.Command) (*rolebindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	resource := ""
	if c.cliName != "ccloud" {
		resource, err = cmd.Flags().GetString("resource")
		if err != nil {
			return nil, errors.HandleCommon(err, cmd)
		}
	}

	prefix := cmd.Flags().Changed("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	scope := &mds.MdsScope{}
	scopeV2 := &mdsv2alpha1.Scope{}
	if c.cliName != "ccloud" {
		scope, err = c.parseAndValidateScope(cmd)
	} else {
		scopeV2, err = c.parseAndValidateScopeV2(cmd)
	}
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	resourcesRequest := mds.ResourcesRequest{}
	if resource != "" {
		parsedResourcePattern, err := c.parseAndValidateResourcePattern(resource, prefix)
		if err != nil {
			return nil, errors.HandleCommon(err, cmd)
		}
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		if err != nil {
			return nil, errors.HandleCommon(err, cmd)
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
		return errors.HandleCommon(err, cmd)
	}

	var resp *http.Response
	if c.cliName == "ccloud" {
		resp, err = c.ccloudCreate(options)
	} else {
		resp, err = c.confluentCreate(options)
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.HandleCommon(errors.Wrapf(err, "No error, but received HTTP status code %d.  Please file a support ticket with details", resp.StatusCode), cmd)
	}

	return nil
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
		return errors.HandleCommon(err, cmd)
	}

	var resp *http.Response
	if c.cliName == "ccloud" {
		resp, err = c.ccloudDelete(options)
	} else {
		resp, err = c.confluentDelete(options)
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.HandleCommon(errors.Wrapf(err, "No error, but received HTTP status code %d.  Please file a support ticket with details", resp.StatusCode), cmd)
	}

	return nil
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
