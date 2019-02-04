package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/codyaray/go-printer"
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
			return common.HandleError(err, cmd)
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

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	resp, err := c.client.ListACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))

	if err != nil {
		return common.HandleError(err, cmd)
	}

	var bindings [][]string
	for _, binding := range resp {

		record := &struct {
			Principal  string
			Permission string
			Operation  string
			Resource   string
			Name       string
		}{
			binding.Entry.Principal,
			binding.Entry.PermissionType.String(),
			binding.Entry.Operation.String(),
			binding.Pattern.ResourceType.String(),
			binding.Pattern.Name,
		}
		bindings = append(bindings, printer.ToRow(record,
			[]string{"Principal", "Permission", "Operation", "Resource", "Name"}))
	}

	printer.RenderCollectionTable(bindings, []string{"Principal", "Permission", "Operation", "Resource", "Name"})

	return nil
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	if acl.errors != nil {
		return common.HandleError(fmt.Errorf("Failed to parse input \n\t"+strings.Join(acl.errors, "\n\t")), cmd)
	}

	err = c.client.CreateACL(context.Background(), cluster, []*kafkav1.ACLBinding{acl.ACLBinding})

	return common.HandleError(err, cmd)
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	if acl.errors != nil {
		return common.HandleError(fmt.Errorf(strings.Join(acl.errors, "\n\t")), cmd)
	}

	cluster, err := common.Cluster(c.config)
	if err != nil {
		return common.HandleError(err, cmd)
	}

	err = c.client.DeleteACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))

	return common.HandleError(err, cmd)
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(binding *ACLConfiguration) *ACLConfiguration {
	if binding.Entry.PermissionType == kafkav1.ACLPermissionTypes_UNKNOWN {
		binding.errors = append(binding.errors, "--allow or --deny must be specified when adding or deleting an acl")
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == kafkav1.ResourceTypes_UNKNOWN {
		binding.errors = append(binding.errors, "a resource flag must be specified when adding or deleting an acl")
	}

	return binding
}

// validateList ensures the basic requirements for acl list are met
func validateList(binding *ACLConfiguration) *ACLConfiguration {
	if binding.Entry.Principal == "" || binding.Pattern == nil {
		binding.errors = append(binding.errors, "either --principal or a resource must be specified when listing acls not both ")
	}
	return binding
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *kafkav1.ACLBinding) *kafkav1.ACLFilter {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113	if binding.Entry == nil {
	if binding.Entry == nil {
		binding.Entry = new(kafkav1.AccessControlEntryConfig)
	}

	if binding.Entry.Operation == kafkav1.ACLOperations_UNKNOWN {
		binding.Entry.Operation = kafkav1.ACLOperations_ANY
	}

	if binding.Entry.PermissionType == kafkav1.ACLPermissionTypes_UNKNOWN {
		binding.Entry.PermissionType = kafkav1.ACLPermissionTypes_ANY
	}

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if binding.Pattern == nil {
		binding.Pattern = &kafkav1.ResourcePatternConfig{}
	}

	binding.Entry.Host = "*"

	if binding.Pattern.ResourceType == kafkav1.ResourceTypes_UNKNOWN {
		binding.Pattern.ResourceType = kafkav1.ResourceTypes_ANY
	}

	if binding.Pattern.PatternType == kafkav1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = kafkav1.PatternTypes_ANY
	}

	return &kafkav1.ACLFilter{
		EntryFilter:   binding.Entry,
		PatternFilter: binding.Pattern,
	}
}
