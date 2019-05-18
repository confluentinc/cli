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

	//"github.com/confluentinc/go-printer"
	"context"

	mds "github.com/confluentinc/mds-sdk-go"
	//"fmt"
)

var (
	// rolebindingListFields     = []string{"Name", "SuperUser", "AllowedOperations"}
	// rolebindingListLabels     = []string{"Name", "SuperUser", "AllowedOperations"}
	// rolebindingDescribeFields = []string{"Name", "SuperUser", "AllowedOperations"}
	// rolebindingDescribeLabels = []string{"Name", "SuperUser", "AllowedOperations"}
	resourcePatternListFields = []string{"Name", "ResourceType", "PatternType"}
	resourcePatternListLabels = []string{"Role", "Name", "ResourceType", "PatternType"}
)

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
			Short: "Manage RBAC/IAM rolebindings",
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
		Short: "List rolebindings",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("principal", "", "Principal whose rolebindings should be listed")
	listCmd.Flags().String("role", "", "List rolebindings under a specific role given to a principal")
	listCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("schema-registry-cluster-id", "", "Schema registry cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for scope of rolebinding listings")
	listCmd.Flags().String("connect-cluster-id", "", "Connect cluster ID for scope of rolebinding listings")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new rolebinding",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("role", "", "Role name of the new rolebinding")
	createCmd.Flags().String("resource", "", "Qualified resource name for the rolebinding")
	createCmd.Flags().Bool("prefix", false, "Whether the provided resource name should be treated as a prefix pattern")
	createCmd.Flags().String("principal", "", "Qualified principal name for the rolebinding")
	createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the rolebinding")
	createCmd.Flags().String("schema-registry-cluster-id", "", "Schema registry cluster ID for the rolebinding")
	createCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the rolebinding")
	createCmd.Flags().String("connect-cluster-id", "", "Connect cluster ID for the rolebinding")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)

	deleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing rolebinding",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().String("role", "", "Role name of the existing rolebinding")
	deleteCmd.Flags().String("resource", "", "Qualified resource name associated with the rolebinding")
	deleteCmd.Flags().String("principal", "", "Qualified principal name associated with the rolebinding")
	deleteCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the rolebinding")
	deleteCmd.Flags().String("schema-registry-cluster-id", "", "Schema registry cluster ID for the rolebinding")
	deleteCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the rolebinding")
	deleteCmd.Flags().String("connect-cluster-id", "", "Connect cluster ID for the rolebinding")
	deleteCmd.Flags().SortFlags = false
	c.AddCommand(deleteCmd)
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
		//fmt.Println(roleNames)

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

func (c *rolebindingCommand) parseResourcePattern(typename string, prefix bool) mds.ResourcePattern {
	var result mds.ResourcePattern
	if prefix {
		result.PatternType = "PREFIXED"
	} else {
		result.PatternType = "LITERAL"
	}

	result.ResourceType = strings.Split(typename, ":")[0]
	result.Name = typename[strings.Index(typename, ":")+1:]

	return result
}

func (c *rolebindingCommand) validateRoleAndResourceType(roleName string, resourceType string) error {
	role, _, err := c.client.RoleDefinitionsApi.RoleDetail(c.ctx, roleName)
	if err != nil {
		return errors.Wrapf(err, "Failed to look up role "+roleName+", maybe invalid role name was specified?")
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
		return errors.New("Invalid resource type " + resourceType + " specified, must be one of " + strings.Join(allResourceTypes, ","))
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

	id, err = cmd.Flags().GetString("schema-registry-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.SchemaRegistryCluster = id

	id, err = cmd.Flags().GetString("ksql-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.KsqlCluster = id

	id, err = cmd.Flags().GetString("connect-cluster-id")
	if err != nil {
		return nil, errors.HandleCommon(err, cmd)
	}
	scope.ConnectCluster = id

	if scope.KafkaCluster == "" && (scope.SchemaRegistryCluster != "" ||
		scope.KsqlCluster != "" ||
		scope.ConnectCluster != "") {
		return nil, errors.HandleCommon(errors.New("A Kafka cluster ID must also be specified to uniquely identify the scope"), cmd)
	}

	if scope.KafkaCluster == "" &&
		scope.SchemaRegistryCluster == "" &&
		scope.KsqlCluster == "" &&
		scope.ConnectCluster == "" {
		return nil, errors.HandleCommon(errors.New("Must specify at least one cluster ID flag to indicate rolebinding scope"), cmd)
	}

	return scope, nil
}

func (c *rolebindingCommand) create(cmd *cobra.Command, args []string) error {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	prefix := cmd.Flags().Changed("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	scopeClusters, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resp := (*http.Response)(nil)
	if resource != "" {
		parsedResourcePattern := c.parseResourcePattern(resource, prefix)
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		rp := []mds.ResourcePattern{
			parsedResourcePattern,
		}
		rr := mds.ResourcesRequest{
			Scope:            mds.Scope{Clusters: *scopeClusters},
			ResourcePatterns: rp,
		}
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleResourcesForPrincipal(c.ctx, principal, role, rr)
	} else {
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleForPrincipal(c.ctx, principal, role, mds.Scope{Clusters: *scopeClusters})
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return errors.HandleCommon(errors.Wrapf(err, "No error but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, args []string) error {
	role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	prefix := cmd.Flags().Changed("prefix")

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	scopeClusters, err := c.parseAndValidateScope(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resp := (*http.Response)(nil)
	if resource != "" {
		parsedResourcePattern := c.parseResourcePattern(resource, prefix)
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		rp := []mds.ResourcePattern{
			parsedResourcePattern,
		}
		rr := mds.ResourcesRequest{
			Scope:            mds.Scope{Clusters: *scopeClusters},
			ResourcePatterns: rp,
		}
		resp, err = c.client.UserAndRoleMgmtApi.RemoveRoleResourcesForPrincipal(c.ctx, principal, role, rr)
	} else {
		resp, err = c.client.UserAndRoleMgmtApi.DeleteRoleForPrincipal(c.ctx, principal, role, mds.Scope{Clusters: *scopeClusters})
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != 200 && resp.StatusCode != 204 {
		return errors.HandleCommon(errors.Wrapf(err, "No error but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}
