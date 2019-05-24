package iam

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/go-printer"

	"context"

	mds "github.com/confluentinc/mds-sdk-go"
)

var (
	resourcePatternListFields = []string{"ResourceType", "Name", "PatternType"}
	resourcePatternListLabels = []string{"Role", "ResourceType", "Name", "PatternType"}
)

type rolebindingOptions struct {
	role             string
	resource         string
	prefix           bool
	principal        string
	scopeClusters    mds.ScopeClusters
	resourcesRequest mds.ResourcesRequest
}

type rolebindingCommand struct {
	*cobra.Command
	config *config.Config
	ch     *pcmd.ConfigHelper
	client *mds.APIClient
	ctx    context.Context
}

// NewRolebindingCommand returns the sub-command object for interacting with RBAC rolebindings.
func NewRolebindingCommand(config *config.Config, ch *pcmd.ConfigHelper, client *mds.APIClient) *cobra.Command {
	cmd := &rolebindingCommand{
		Command: &cobra.Command{
			Use:   "rolebinding",
			Short: "Manage RBAC and IAM role bindings",
			Long: "Manage Role Based Access (RBAC) and Identity and Access Management (IAM) role bindings.",
		},
		config: config,
		ch:     ch,
		client: client,
		ctx:    context.WithValue(context.Background(), mds.ContextAccessToken, config.AuthToken),
	}

	cmd.init()
	return cmd.Command
}

func (c *rolebindingCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List role bindings",
		Long: "List the role bindings for a particular principal and scope.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("principal", "", "Principal whose rolebindings should be listed")
	listCmd.Flags().String("role", "", "List rolebindings under a specific role given to a principal")
	listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for scope of rolebinding listings")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new role binding",
		Long: "Create a new role binding.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("role", "", "Role name of the new role binding.")
	createCmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	createCmd.Flags().Bool("prefix", false, `Whether the provided resource name is treated as a
prefix pattern. The default is false.`)
	createCmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	createCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	createCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
	createCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing role binding",
		Long: "Delete an existing role binding.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("role", "", "Role name of the existing role binding.")
	deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the role binding.")
	deleteCmd.Flags().Bool("prefix", false, `Whether the provided resource name is treated as a
prefix pattern. The default is false.`)
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
	deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
	deleteCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the role binding.")
	deleteCmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
	deleteCmd.Flags().SortFlags = false
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

	if len(strings.Split(typename, ":")) == 1 {
		return result, errors.New("Resource must be specified in this format: <Resource Type>:<Resource Name>")
	}

	result.ResourceType = strings.Split(typename, ":")[0]
	result.Name = typename[strings.Index(typename, ":")+1:]

	return result, nil
}

func (c *rolebindingCommand) validateRoleAndResourceType(roleName string, resourceType string) error {
	role, _, err := c.client.RoleDefinitionsApi.RoleDetail(c.ctx, roleName)
	if err != nil {
		return errors.Wrapf(err, "Failed to look up role "+roleName+". Was an invalid role name specified?")
	}

	allResourceTypes := []string{}
	found := false
	for _, operation := range role.AccessPolicy.AllowedOperations {
		allResourceTypes = append(allResourceTypes, operation.ResourceType)
		if operation.ResourceType == resourceType {
			found = true
		}
	}

	if !found {
		return errors.New("Invalid resource type " + resourceType + " specified. It must be one of " + strings.Join(allResourceTypes, ","))
	}

	return nil
}

func (c *rolebindingCommand) parseAndValidateScope(cmd *cobra.Command) (*mds.ScopeClusters, error) {
	scope := &mds.ScopeClusters{}

	id, err := cmd.Flags().GetString("kafka-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.KafkaCluster = id

	nonKafkaScopesSet := 0

	if cmd.Flags().Changed("schema-registry-cluster-id") {
		nonKafkaScopesSet++
	}
	id, err = cmd.Flags().GetString("schema-registry-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.SchemaRegistryCluster = id

	if cmd.Flags().Changed("ksql-cluster-id") {
		nonKafkaScopesSet++
	}
	id, err = cmd.Flags().GetString("ksql-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.KsqlCluster = id

	if cmd.Flags().Changed("connect-cluster-id") {
		nonKafkaScopesSet++
	}
	id, err = cmd.Flags().GetString("connect-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.ConnectCluster = id

	if scope.KafkaCluster == "" && (scope.SchemaRegistryCluster != "" ||
		scope.KsqlCluster != "" ||
		scope.ConnectCluster != "") {
		return nil, errors.HandleCommon(errors.New("Must also specify a Kafka cluster ID to uniquely identify the scope."), cmd)
	}

	if scope.KafkaCluster == "" &&
		scope.SchemaRegistryCluster == "" &&
		scope.KsqlCluster == "" &&
		scope.ConnectCluster == "" {
		return nil, errors.HandleCommon(errors.New("Must specify at least one cluster ID flag to indicate role binding scope."), cmd)
	}

	if nonKafkaScopesSet > 1 {
		return nil, errors.HandleCommon(errors.New("Cannot specify more than one non-Kafka cluster ID for a scope."), cmd)
	}

	return scope, nil
}

func (c *rolebindingCommand) list(cmd *cobra.Command, args []string) error {
	role := "*"
	if cmd.Flags().Changed("role") {
		r, err := cmd.Flags().GetString("role")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		role = r
	}

	principal, err := cmd.Flags().GetString("principal")
	if err == nil && principal == "" {
		return errors.New("You must specify a principal to list role bindings for.")
	}
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	scopeClusters, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	var roleNamesWithMultiplicity []string
	var resourcePatterns []mds.ResourcePattern
	if role == "*" {
		roleNames, _, err := c.client.UserAndRoleMgmtApi.ScopedPrincipalRolenames(c.ctx, principal, mds.Scope{Clusters: *scopeClusters})
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		for _, r := range roleNames {
			rps, _, err := c.client.UserAndRoleMgmtApi.GetRoleResourcesForPrincipal(c.ctx, principal, r, mds.Scope{Clusters: *scopeClusters})
			if len(rps) == 0 {
				rps = []mds.ResourcePattern{mds.ResourcePattern{}}
			}
			resourcePatterns = append(resourcePatterns, rps...)
			for range rps {
				roleNamesWithMultiplicity = append(roleNamesWithMultiplicity, r)
			}
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
		}
	} else {
		resourcePatterns, _, err = c.client.UserAndRoleMgmtApi.GetRoleResourcesForPrincipal(c.ctx, principal, role, mds.Scope{Clusters: *scopeClusters})
		if len(resourcePatterns) == 0 {
			resourcePatterns = []mds.ResourcePattern{mds.ResourcePattern{}}
		}
		for range resourcePatterns {
			roleNamesWithMultiplicity = append(roleNamesWithMultiplicity, role)
		}
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	var data [][]string
	for i, pattern := range resourcePatterns {
		data = append(data, append([]string{roleNamesWithMultiplicity[i]}, printer.ToRow(&pattern, resourcePatternListFields)...))
	}
	printer.RenderCollectionTable(data, resourcePatternListLabels)

	return nil
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
	if err == nil && principal == "" {
		return nil, errors.New("Must specify a principal for the role binding.")
	}
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	err = c.validatePrincipalFormat(principal)
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}

	scopeClusters, err := c.parseAndValidateScope(cmd)
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
			Scope:            mds.Scope{Clusters: *scopeClusters},
			ResourcePatterns: resourcePatterns,
		}
	}

	return &rolebindingOptions{
			role,
			resource,
			prefix,
			principal,
			*scopeClusters,
			resourcesRequest,
		},
		nil
}

func (c *rolebindingCommand) create(cmd *cobra.Command, args []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resp := (*http.Response)(nil)
	var addErr error
	if options.resource != "" {
		resp, addErr = c.client.UserAndRoleMgmtApi.AddRoleResourcesForPrincipal(c.ctx, options.principal, options.role, options.resourcesRequest)
	} else {
		resp, addErr = c.client.UserAndRoleMgmtApi.AddRoleForPrincipal(c.ctx, options.principal, options.role, mds.Scope{Clusters: options.scopeClusters})
	}

	if addErr != nil {
		return errors.HandleCommon(addErr, cmd)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return errors.HandleCommon(errors.Wrapf(err, "No error, but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, args []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resp := (*http.Response)(nil)
	var removeErr error
	if options.resource != "" {
		resp, removeErr = c.client.UserAndRoleMgmtApi.RemoveRoleResourcesForPrincipal(c.ctx, options.principal, options.role, options.resourcesRequest)
	} else {
		resp, removeErr = c.client.UserAndRoleMgmtApi.DeleteRoleForPrincipal(c.ctx, options.principal, options.role, mds.Scope{Clusters: options.scopeClusters})
	}

	if removeErr != nil {
		return errors.HandleCommon(removeErr, cmd)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return errors.HandleCommon(errors.Wrapf(err, "No error, but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}
