package rbac

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/config"
	//"github.com/confluentinc/cli/internal/pkg/errors"
	//"github.com/confluentinc/go-printer"
	mds "github.com/confluentinc/mds-sdk-go"
	//"context"
	//"fmt"
	//"strings"
)

var (
	resourceListFields     = []string{"Name", "SuperUser", "AllowedOperations"}
	resourceListLabels     = []string{"Name", "SuperUser", "AllowedOperations"}
	resourceDescribeFields = []string{"Name", "SuperUser", "AllowedOperations"}
	resourceDescribeLabels = []string{"Name", "SuperUser", "AllowedOperations"}
)

type resourcesCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
}

// NewResourcesCommand returns the sub-command object for interacting with RBAC resources.
func NewResourcesCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &resourcesCommand{
		Command: &cobra.Command{
			Use:   "resource",
			Short: "Manage RBAC/IAM resources",
		},
		config: config,
		client: client,
	}

	cmd.init()
	return cmd.Command
}

func (c *resourcesCommand) init() {
	c.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List permissions current user has across all resources",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	})
}

func (c *resourcesCommand) list(cmd *cobra.Command, args []string) error {
	// TODO there is no /security/1.0/resources?

	return nil
}
