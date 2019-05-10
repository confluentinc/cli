package iam

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"

	//"github.com/confluentinc/go-printer"
	"context"

	mds "github.com/confluentinc/mds-sdk-go"
	//"fmt"
)

var (
	rolebindingListFields     = []string{"Name", "SuperUser", "AllowedOperations"}
	rolebindingListLabels     = []string{"Name", "SuperUser", "AllowedOperations"}
	rolebindingDescribeFields = []string{"Name", "SuperUser", "AllowedOperations"}
	rolebindingDescribeLabels = []string{"Name", "SuperUser", "AllowedOperations"}
)

type rolebindingCommand struct {
	*cobra.Command
	config *config.Config
	ch     *pcmd.ConfigHelper
	client *mds.APIClient
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
	listCmd.Flags().String("cluster", "", "Cluster ID for which to list rolebindings")
	listCmd.Flags().String("topic", "", "Topic name for which to list rolebindings")
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
	c.AddCommand(deleteCmd)
}

func (c *rolebindingCommand) list(cmd *cobra.Command, args []string) error {
	// TODO see https://confluent.slack.com/archives/CFA95LYAK/p1554844878211100

	// rolebindings, _, err := c.client.ControlPlaneApi.rolebindings(context.Background())
	// if err != nil {
	// 	return errors.HandleCommon(err, cmd)
	// }

	// var data [][]string
	// for _, rolebinding := range rolebindings {
	// 	data = append(data, printer.ToRow(&rolebinding, rolebindingListFields))
	// }
	// printer.RenderCollectionTable(data, listLabels)

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
	fmt.Println("GET ROLE DETAIL " + roleName)
	role, _, err := c.client.RoleDefinitionsApi.RoleDetail(context.Background(), roleName)
	if err != nil {
		return errors.Wrapf(err, "Failed to look up role "+roleName+", maybe invalid role name was specified?")
	}
	fmt.Println("ROLE")
	fmt.Println(role)

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

	resp, err := (*http.Response)(nil), (error)(nil)
	if resource != "" {
		fmt.Println("HERE2")
		parsedResourcePattern := c.parseResourcePattern(resource, prefix)
		fmt.Println("HERE3")
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		fmt.Println("HERE4")
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
		fmt.Println("HERE5")
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleResourcesForPrincipal(context.Background(), principal, role, rr)
	} else {
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleForPrincipal(context.Background(), principal, role, mds.Scope{Clusters: *scopeClusters})
	}

	fmt.Println(resp)
	fmt.Println(resp.StatusCode)
	fmt.Println(err)

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != 200 {
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

	resp, err := (*http.Response)(nil), (error)(nil)
	if resource != "" {
		fmt.Println("HERE2")
		parsedResourcePattern := c.parseResourcePattern(resource, prefix)
		fmt.Println("HERE3")
		err = c.validateRoleAndResourceType(role, parsedResourcePattern.ResourceType)
		fmt.Println("HERE4")
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
		fmt.Println("HERE5")
		resp, err = c.client.UserAndRoleMgmtApi.RemoveRoleResourcesForPrincipal(context.Background(), principal, role, rr)
	} else {
		resp, err = c.client.UserAndRoleMgmtApi.DeleteRoleForPrincipal(context.Background(), principal, role, mds.Scope{Clusters: *scopeClusters})
	}

	fmt.Println(resp)
	fmt.Println(resp.StatusCode)
	fmt.Println(err)

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != 200 {
		return errors.HandleCommon(errors.Wrapf(err, "No error but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}
