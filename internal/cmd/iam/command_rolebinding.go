package iam

import (
	"net/http"
	"strconv"

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
			Use:   "Rolebinding",
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
	createCmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the rolebinding")
	createCmd.Flags().String("schema-registry-cluster-id", "", "Schema registry cluster ID for the rolebinding")
	createCmd.Flags().String("ksql-cluster-id", "", "KSQL cluster ID for the rolebinding")
	createCmd.Flags().String("connect-cluster-id", "", "Connect cluster ID for the rolebinding")
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
		return nil, errors.HandleCommon(errors.Wrapf(err, "A Kafka cluster ID must also be specified to uniquely identify the scope"), cmd)
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
		rp := []mds.ResourcePattern{}
		rr := mds.ResourcesRequest{
			Scope:            mds.Scope{Clusters: *scopeClusters},
			ResourcePatterns: rp, // TODO
		}
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleResourcesForPrincipal(context.Background(), principal, role, rr)
	} else {
		resp, err = c.client.UserAndRoleMgmtApi.AddRoleForPrincipal(context.Background(), principal, role, mds.Scope{Clusters: *scopeClusters})
	}

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if resp.StatusCode != 200 {
		return errors.HandleCommon(errors.Wrapf(err, "No error but received HTTP status code "+strconv.Itoa(resp.StatusCode)), cmd)
	}

	return nil
}

func (c *rolebindingCommand) delete(cmd *cobra.Command, args []string) error {
	/*role, err := cmd.Flags().GetString("role")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	resource, err := cmd.Flags().GetString("resource")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	principal, err := cmd.Flags().GetString("principal")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	cluster, err := pcmd.GetKafkaCluster(cmd, c.ch)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}*/

	// TODO https://confluent.slack.com/archives/CFA95LYAK/p1554844565210200
	/*success, _, err := c.client.ControlPlaneApi.DeleteRoleForPrincipal(context.Background(), principal, cluster.Id+resource, role)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if !success {
		return errors.HandleCommon(errors.New("the server received a valid request but did not delete the rolebinding -- perhaps parameters are invalid?"), cmd)
	}*/

	return nil
}
