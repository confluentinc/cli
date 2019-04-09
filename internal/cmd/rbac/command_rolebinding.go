package rbac

import (
	"github.com/spf13/cobra"

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

type rolebindingsCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// NewRolebindingsCommand returns the sub-command object for interacting with RBAC rolebindings.
func NewRolebindingsCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &rolebindingsCommand{
		Command: &cobra.Command{
			Use:   "Rolebinding",
			Short: "Manage RBAC/IAM rolebindings",
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *rolebindingsCommand) init() {
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
	createCmd.Flags().String("cluster", "", "Cluster ID for the rolebinding")
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
	deleteCmd.Flags().String("cluster", "", "Cluster ID of the rolebinding")
	c.AddCommand(deleteCmd)
}

func (c *rolebindingsCommand) list(cmd *cobra.Command, args []string) error {
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

func (c *rolebindingsCommand) create(cmd *cobra.Command, args []string) error {
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

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// TODO https://confluent.slack.com/archives/CFA95LYAK/p1554844565210200
	success, _, err := c.client.ControlPlaneApi.AddRoleForPrincipal(context.Background(), principal, cluster+resource, role)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if !success {
		return errors.HandleCommon(errors.New("the server received a valid request but did not create the rolebinding -- perhaps parameters are invalid?"), cmd)
	}

	return nil
}

func (c *rolebindingsCommand) delete(cmd *cobra.Command, args []string) error {
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

	cluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	// TODO https://confluent.slack.com/archives/CFA95LYAK/p1554844565210200
	success, _, err := c.client.ControlPlaneApi.DeleteRoleForPrincipal(context.Background(), principal, cluster+resource, role)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	if !success {
		return errors.HandleCommon(errors.New("the server received a valid request but did not delete the rolebinding -- perhaps parameters are invalid?"), cmd)
	}

	return nil
}
