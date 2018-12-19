package kafka

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	chttp "github.com/confluentinc/ccloud-sdk-go"
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
)

type aclCommand struct {
	*cobra.Command
	config *shared.Config
	client chttp.Kafka
}

// NewACLCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewACLCommand(config *shared.Config, plugin common.Provider) *cobra.Command {
	cmd := &aclCommand{
		Command: &cobra.Command{
			Use:   "acl",
			Short: "Manage Kafka ACLs.",
		},
		config: config,
	}

	cmd.init(plugin)
	return cmd.Command
}

func (c *aclCommand) init(plugin common.Provider) {

	c.Command.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if err := c.config.CheckLogin(); err != nil {
			return err
		}
		// Lazy load plugin to avoid unnecessarily spawning child processes
		return plugin(&c.client)
	}

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(aclConfigFlags())
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(aclConfigFlags())
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(resourceFlags())
	cmd.Flags().String("principal", "*", "Set ACL filter principal")
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl := validateList(parse(cmd))

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	resp, err := c.client.ListACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))
	if err != nil {
		return common.HandleError(err, cmd)
	}

	jsonPrinter.PrintObj(resp, os.Stdout)

	return nil
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	if acl.errors != nil {
		return fmt.Errorf("Failed to process input\n\t%v", strings.Join(acl.errors, "\n\t"))
	}

	err = c.client.CreateACL(context.Background(), cluster, []*kafkav1.ACLBinding{acl.ACLBinding})
	if err != nil {
		return common.HandleError(err, cmd)
	}

	return nil
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	cluster, err := c.currentCluster()
	if err != nil {
		return err
	}

	if acl.errors != nil {
		return fmt.Errorf("Failed to process input\n\t%v", strings.Join(acl.errors, "\n\t"))
	}

	err = c.client.DeleteACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return nil
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(b *ACLConfiguration) *ACLConfiguration {
	if b.Entry.PermissionType == kafkav1.ACLPermissionTypes_ANY {
		b.errors = append(b.errors, "--allow or --deny must be specified when adding or deleting an acl")
	}

	if b.Entry.Principal == "" {
		b.errors = append(b.errors, "--principal must be specified when adding or deleting acls")
	}

	if b.Entry.Operation == kafkav1.ACLOperations_ANY {
		b.errors = append(b.errors, "--operation must be specified when adding or deleting acls")
	}
	if b.Pattern == nil || b.Pattern.ResourceType == kafkav1.ResourceTypes_ANY {
		b.errors = append(b.errors, "a resource flag must be specified when adding or deleting an acl")
	}
	return b
}

// validateList ensures the basic requirements for acl list are met
func validateList(b *ACLConfiguration) *ACLConfiguration {
	if b.Entry.Principal == "" || b.Pattern == nil {
		b.errors = append(b.errors,
			"either --principal or a resource must be specified when listing acls not both ")
	}
	return b
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *kafkav1.ACLBinding) *kafkav1.ACLFilter {
	if binding.Entry == nil {
		binding.Entry = new(kafkav1.AccessControlEntryConfig)
	}

	if binding.Pattern == nil {
		binding.Pattern = &kafkav1.ResourcePatternConfig{}
		binding.Pattern.Name = "*"
	}
	binding.Entry.Host = "*"

	if binding.Entry.Principal == "" {
		binding.Entry.Principal = "User:*"
	}

	filter := &kafkav1.ACLFilter{
		EntryFilter:   binding.Entry,
		PatternFilter: binding.Pattern,
	}
	return filter
}

// currentCluster returns the current cluster context
func (c *aclCommand) currentCluster() (*kafkav1.Cluster, error) {
	ctx, err := c.config.Context()
	if ctx == nil {
		return nil, fmt.Errorf("no cluster context is currently selected")
	}

	conf, err := c.config.KafkaClusterConfig()
	if err != nil {
		return nil, err
	}

	return &kafkav1.Cluster{AccountId: c.config.Auth.Account.Id, Id: ctx.Kafka, ApiEndpoint: conf.APIEndpoint}, nil
}
