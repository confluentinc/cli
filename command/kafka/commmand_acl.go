package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/command/common"
	"github.com/confluentinc/cli/shared"
	"github.com/confluentinc/cli/shared/kafka"
	"github.com/spf13/cobra"
	"strings"
	"os"
)

type aclCommand struct {
	*cobra.Command
	config *shared.Config
}

// NewACLCommand returns the Cobra clusterCommand for Kafka Cluster.
func NewACLCommand(config *shared.Config) *cobra.Command {
	cmd := &aclCommand{
		Command: &cobra.Command{
			Use:   "acl",
			Short: "Manage Kafka ACLs.",
		},
		config: config,
	}

	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(ACLConfigFlags())
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(ACLConfigFlags())
	cmd.Flags().SortFlags = false
	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(ResourceFlags())
	cmd.Flags().String("principal", "*", "Set ACL filter principal")
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl := validateList(parse(cmd))
	if acl.errors != nil {
		return fmt.Errorf("Failed to process input\n\t %s", acl.errors)
	}

	resp, err := Client.ListACL(context.Background(), convertToFilter(acl.KafkaAPIACLRequest))
	if err != nil {
		return common.HandleError(err, cmd)
	}

	jsonPrinter.PrintObj(resp.Results, os.Stdout)

	return nil
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))
	if acl.errors != nil {
		return fmt.Errorf("Failed to process input\n\t%v", strings.Join(acl.errors, "\n\t"))
	}
	_, err := Client.CreateACL(context.Background(), acl.KafkaAPIACLRequest)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	return nil
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))
	if acl.errors != nil {
		return fmt.Errorf("Failed to process input\n\t%v", strings.Join(acl.errors, "\n\t"))
	}
	_, err := Client.DeleteACL(context.Background(), convertToFilter(acl.KafkaAPIACLRequest))
	if err != nil {
		return common.HandleError(err, cmd)
	}
	return nil
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(b *ACLConfiguration) *ACLConfiguration {
	if !common.IsSet(b.Entry.PermissionType) {
		b.errors = append(b.errors, "--allow or --deny must be specified when adding or deleting an acl")
	}
	if !common.IsSet(b.Entry.Principal) {
		b.errors = append(b.errors, "--principal must be specified when adding or deleting acls")
	}
	if !common.IsSet(b.Entry.Operation) {
		b.errors = append(b.errors, "--operation must be specified when adding or deleting acls")
	}
	if b.Pattern == nil || !common.IsSet(b.Pattern.ResourceType) {
		b.errors = append(b.errors, "a resource flag must be specified when adding or deleting an acl")
	}
	return b
}

// validateList ensures the basic requirements for acl list are met
func validateList(b *ACLConfiguration) *ACLConfiguration {
	if !common.IsSet(b.Entry.Principal) && !common.IsSet(b.Pattern) {
		b.errors = append(b.errors,
			"either --principal or a resource must be specified when listing acls not both ")
	}
	return b
}

// convertToFilter converts a KafkaAPIACLRequest to a KafkaAPIACLFilterRequest
func convertToFilter(b *kafka.KafkaAPIACLRequest) *kafka.KafkaAPIACLFilterRequest {
	b.Entry.Operation = kafka.AccessControlEntryConfig_ANY.String()
	b.Entry.PermissionType = kafka.AccessControlEntryConfig_ANY.String()
	b.Entry.Host = ""

	if !common.IsSet(b.Pattern) {
		b.Pattern = &kafka.ResourcePatternConfig{}
		b.Pattern.ResourceType = kafka.ResourcePatternConfig_ANY.String()
		b.Pattern.PatternType = kafka.ResourcePatternConfig_ANY.String()
	}

	return &kafka.KafkaAPIACLFilterRequest{
		EntryFilter: b.Entry,
		PatternFilter: b.Pattern,
	}
}
