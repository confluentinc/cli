package iam

import (
	"context"
	"fmt"
	"github.com/confluentinc/mds-sdk-go/mdsv2alpha1"
	"net/http"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
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
)

type rolebindingOptions struct {
	role             string
	resource         string
	prefix           bool
	principal        string
	scopeV2			 mdsv2alpha1.Scope
	mdsScope         mds.MdsScope
	resourcesRequest mds.ResourcesRequest
}

type rolebindingCommand struct {
	*cmd.AuthenticatedCLICommand
}

type listDisplay struct {
	Principal    string
	Role         string
	ResourceType string
	Name         string
	PatternType  string
}

// NewRolebindingCommand returns the sub-command object for interacting with RBAC rolebindings.
func NewRolebindingCommand(cfg *v3.Config, prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "rolebinding",
			Short: "Manage RBAC and IAM role bindings.",
			Long:  "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) role bindings.",
		},
		cfg, prerunner)
	roleBindingCmd := &rolebindingCommand{AuthenticatedCLICommand: cliCmd}
	roleBindingCmd.init()
	return roleBindingCmd.Command
}

func (c *rolebindingCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List role bindings.",
		Long:  "List the role bindings for a particular principal and/or role, and a particular scope.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("principal", "", "Principal whose rolebindings should be listed.")
	listCmd.Flags().String("role", "", "List rolebindings under a specific role given to a principal. Or if no principal is specified, list principals with the role.")
	listCmd.Flags().String("resource", "", "If specified with a role and no principals, list principals with rolebindings to the role for this qualified resource.")
	listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
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
	createCmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	createCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	createCmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	createCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	createCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
	createCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	createCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
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
	deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the role binding.")
	deleteCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	deleteCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
	deleteCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	deleteCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
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
	if c.Config.CLIName == "ccloud" {
		return nil
	}
	ctx := c.createContext()
	role, _, err := c.MDSClient.RBACRoleDefinitionsApi.RoleDetail(ctx, roleName)
	if err != nil {
		return errors.Wrapf(err, "Failed to look up role %s. Was an invalid role name specified?", roleName)
	}

	allResourceTypes := []string{}
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
	scope := &mds.ScopeClusters{}
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
		if c.Config.CLIName != "ccloud" {
			if scope.KafkaCluster == "" && nonKafkaScopesSet > 0 {
				return nil, errors.HandleCommon(errors.New("Must also specify a --kafka-cluster-id to uniquely identify the scope."), cmd)
			}

			if scope.KafkaCluster == "" && nonKafkaScopesSet == 0 {
				return nil, errors.HandleCommon(errors.New("Must specify either cluster ID flag to indicate role binding scope or the cluster name."), cmd)
			}
		}

		if nonKafkaScopesSet > 1 {
			return nil, errors.HandleCommon(errors.New("Cannot specify more than one non-Kafka cluster ID for a scope."), cmd)
		}
		return &mds.MdsScope{Clusters: *scope}, nil
	}

	return &mds.MdsScope{ClusterName: clusterName}, nil
}

func (c *rolebindingCommand) parseAndValidateScopeV2(cmd *cobra.Command) (*mdsv2alpha1.Scope, error) {
	mdsScope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return nil, err
	}
	orgResourceId := c.State.Auth.Organization.GetResourceId()
	scopeV2 := &mdsv2alpha1.Scope{
		Path: []string{"organization=" + orgResourceId, "environment=" + c.EnvironmentId()},
		Clusters: mdsv2alpha1.ScopeClusters{
			KafkaCluster: mdsScope.Clusters.KafkaCluster,
			ConnectCluster: mdsScope.Clusters.ConnectCluster,
			KsqlCluster: mdsScope.Clusters.KsqlCluster,
			SchemaRegistryCluster: mdsScope.Clusters.SchemaRegistryCluster,
		},
	}
	return scopeV2, err
}

func (c *rolebindingCommand) list(cmd *cobra.Command, args []string) error {
	if cmd.Flags().Changed("principal") {
		return c.listPrincipalResources(cmd)
	} else if cmd.Flags().Changed("role") {
		return c.listRolePrincipals(cmd)
	}
	return errors.HandleCommon(fmt.Errorf("required: either principal or role is required"), cmd)
}

func (c *rolebindingCommand) confluentListPrincipalResourcesHelper(cmd *cobra.Command, principal string, scope *mds.MdsScope, role string) error {
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

func (c *rolebindingCommand) ccloudListPrincipalResourcesHelper(cmd *cobra.Command, principal string, scopeV2 *mdsv2alpha1.Scope, role string) error {
	scopedPrincipalsRolesResourcePatterns, _, err := c.MDSv2Client.RBACRoleBindingSummariesApi.MyRoleBindings(
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

	for _, scopeRoleBindingMapping := range scopedPrincipalsRolesResourcePatterns {
		// roleBindingScope := scopeRoleBindingMapping.Scope
		principalsRolesResourcePatterns := scopeRoleBindingMapping.Rolebindings
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
	}

	outputWriter.StableSort()

	return outputWriter.Out()
}

func (c *rolebindingCommand) listPrincipalResources(cmd *cobra.Command) error {
	scope := &mds.MdsScope{}
	scopeV2 := &mdsv2alpha1.Scope{}
	var err error
	if c.Config.CLIName == "ccloud" {
		scopeV2, err = c.parseAndValidateScopeV2(cmd)
	} else {
		scope, err = c.parseAndValidateScope(cmd)
	}
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
	if c.Config.CLIName == "ccloud" {
		return c.ccloudListPrincipalResourcesHelper(cmd, principal, scopeV2, role)
	} else {
		return c.confluentListPrincipalResourcesHelper(cmd, principal, scope, role)
	}
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

func (c *rolebindingCommand) listRolePrincipals(cmd *cobra.Command) error {
	scope := &mds.MdsScope{}
	// scopeV2 := &mdsv2alpha1.Scope{}
	var err error
	if c.Config.CLIName == "ccloud" {
		// scopeV2, err = c.parseAndValidateScopeV2(cmd)
	} else {
		scope, err = c.parseAndValidateScope(cmd)
	}
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var principals []string
	if cmd.Flags().Changed("resource") && c.Config.CLIName != "ccloud" {
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

func (c *rolebindingCommand) parseCommon(cmd *cobra.Command) (*rolebindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
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
	if c.Config.CLIName != "ccloud" {
		scope, err = c.parseAndValidateScope(cmd)
	} else {
		scopeV2, err = c.parseAndValidateScopeV2(cmd)
	}
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	resourcesRequest := mds.ResourcesRequest{}
	if c.Config.CLIName != "ccloud" {
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
				MdsScope:         *scope,
				ResourcePatterns: resourcePatterns,
			}
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

func (c *rolebindingCommand) confluentCreateHelper(options *rolebindingOptions) (resp *http.Response, err error) {
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

func (c *rolebindingCommand) ccloudCreateHelper(options *rolebindingOptions) (resp *http.Response, err error) {
	/*if options.resource != "" {
		resp, err = c.MDSv2Client.RBACRoleBindingCRUDApi.AddRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else {*/
		resp, err = c.MDSv2Client.RBACRoleBindingCRUDApi.AddRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			// options.mdsScope)
			options.scopeV2)
	// }
	return
}

func (c *rolebindingCommand) create(cmd *cobra.Command, args []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var resp *http.Response
	if c.Config.CLIName == "ccloud" {
		resp, err = c.ccloudCreateHelper(options)
	} else {
		resp, err = c.confluentCreateHelper(options)
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.HandleCommon(errors.Wrapf(err, "No error, but received HTTP status code %d.  Please file a support ticket with details", resp.StatusCode), cmd)
	}

	return nil
}

func (c* rolebindingCommand) confluentDeleteHelper(options *rolebindingOptions) (resp *http.Response, err error) {
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

func (c* rolebindingCommand) ccloudDeleteHelper(options *rolebindingOptions) (resp *http.Response, err error) {
	/* if options.resource != "" {
		resp, err = c.MDSv2Client.RBACRoleBindingCRUDApi.RemoveRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else { */
		resp, err = c.MDSv2Client.RBACRoleBindingCRUDApi.DeleteRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			// options.mdsScope)
			options.scopeV2)
	// }
	return
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, args []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var resp *http.Response
	if c.Config.CLIName == "ccloud" {
		resp, err = c.ccloudDeleteHelper(options)
	} else {
		resp, err = c.confluentDeleteHelper(options)
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
	if c.Config.CLIName == "ccloud" {
		return context.WithValue(context.Background(), mdsv2alpha1.ContextAccessToken, c.AuthToken())
	} else {
		return context.WithValue(context.Background(), mds.ContextAccessToken, c.AuthToken())
	}
}
