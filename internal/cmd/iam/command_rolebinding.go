package iam

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
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
)

type rolebindingOptions struct {
	role             string
	resource         string
	prefix           bool
	principal        string
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
func NewRolebindingCommand(prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "rolebinding",
			Short: "Manage RBAC and IAM role bindings.",
			Long:  "Manage Role-Based Access Control (RBAC) and Identity and Access Management (IAM) role bindings.",
		}, prerunner)
	roleBindingCmd := &rolebindingCommand{AuthenticatedCLICommand: cliCmd}
	roleBindingCmd.init()
	return roleBindingCmd.Command
}

func (c *rolebindingCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Args:  cobra.NoArgs,
		RunE:  cmd.NewCLIRunE(c.list),
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
	listCmd.Flags().String("resource", "", "If specified with a role and no principals, list principals with rolebindings to the role for this qualified resource.")
	listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of rolebinding listings.")
	listCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role binding.",
		RunE:  cmd.NewCLIRunE(c.create),
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("role", "", "Role name of the new role binding.")
	createCmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	createCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	createCmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	createCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	createCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
	createCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	createCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
	createCmd.Flags().SortFlags = false
	check(createCmd.MarkFlagRequired("role"))
	check(createCmd.MarkFlagRequired("principal"))
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing role binding.",
		RunE:  cmd.NewCLIRunE(c.delete),
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("role", "", "Role name of the existing role binding.")
	deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the role binding.")
	deleteCmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	deleteCmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
	deleteCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	deleteCmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for rolebinding listings.")
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
		return errors.NewWrapErrorWithSuggestions(err, fmt.Sprintf(errors.LookUpRoleErrorMsg, roleName), errors.LookUpRoleSuggestions)
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

func (c *rolebindingCommand) list(cmd *cobra.Command, _ []string) error {
	if cmd.Flags().Changed("principal") {
		return c.listPrincipalResources(cmd)
	} else if cmd.Flags().Changed("role") {
		return c.listRolePrincipals(cmd)
	}
	return errors.New(errors.PrincipalOrRoleRequiredErrorMsg)
}

func (c *rolebindingCommand) listPrincipalResources(cmd *cobra.Command) error {
	scope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return err
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return err
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return err
	}

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

func (c *rolebindingCommand) listRolePrincipals(cmd *cobra.Command) error {
	scope, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return err
	}

	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return err
	}

	var principals []string
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

func (c *rolebindingCommand) parseCommon(cmd *cobra.Command) (*rolebindingOptions, error) {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return nil, err
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return nil, err
	}

	prefix := cmd.Flags().Changed("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return nil, err
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return nil, err
	}

	scope, err := c.parseAndValidateScope(cmd)
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
			*scope,
			resourcesRequest,
		},
		nil
}

func (c *rolebindingCommand) create(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	var resp *http.Response
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

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, resp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	return nil
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	var resp *http.Response
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

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, resp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func (c *rolebindingCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.AuthToken())
}
