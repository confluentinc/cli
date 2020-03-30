package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	createCmd *cobra.Command
	deleteCmd *cobra.Command
	listCmd   *cobra.Command
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewACLCommand returns the Cobra command for Kafka ACL.
func NewACLCommand(prerunner pcmd.PreRunner, config *v3.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "acl",
			Short: `Manage Kafka ACLs.`,
		},
		config, prerunner)
	cmd := &aclCommand{AuthenticatedCLICommand: cliCmd}
	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	c.Command.PersistentFlags().String("cluster", "", "Kafka cluster ID.")

	createCmd = &cobra.Command{
		Use:   "create",
		Short: `Create a Kafka ACL.`,
		Example: `You can only specify one of these flags per command invocation: ` + "``cluster``, ``consumer-group``" + `,
` + "``topic``, or ``transactional-id``" + ` per command invocation. For example, if you want to specify both
` + "``consumer-group`` and ``topic``" + `, you must specify this as two separate commands:

::

	ccloud kafka acl create --allow --service-account 1522 --operation READ --consumer-group \
	java_example_group_1

::

	ccloud kafka acl create --allow --service-account 1522 --operation READ --topic '*'

`,
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	createCmd.Flags().AddFlagSet(aclConfigFlags())
	createCmd.Flags().SortFlags = false

	c.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: `Delete a Kafka ACL.`,
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	deleteCmd.Flags().AddFlagSet(aclConfigFlags())
	deleteCmd.Flags().SortFlags = false

	c.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: `List Kafka ACLs for a resource.`,
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().AddFlagSet(resourceFlags())
	listCmd.Flags().Int("service-account", 0, "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)
}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return aclutil.PrintAcls(cmd, resp, os.Stdout)
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	bindings := []*kafkav1.ACLBinding{}
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return errors.HandleCommon(acl.errors, cmd)
		}
		bindings = append(bindings, acl.ACLBinding)
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.Client.Kafka.CreateACLs(context.Background(), cluster, bindings)

	return errors.HandleCommon(err, cmd)
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	filters := []*kafkav1.ACLFilter{}
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return errors.HandleCommon(acl.errors, cmd)
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.Client.Kafka.DeleteACLs(context.Background(), cluster, filters)

	return errors.HandleCommon(err, cmd)
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(binding *ACLConfiguration) {
	if binding.Entry.PermissionType == kafkav1.ACLPermissionTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf("--allow or --deny must be set when adding or deleting an ACL"))
	}

	if binding.Pattern.PatternType == kafkav1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = kafkav1.PatternTypes_LITERAL
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == kafkav1.ResourceTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf("exactly one of %v must be set",
			listEnum(kafkav1.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"})))
	}
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *kafkav1.ACLBinding) *kafkav1.ACLFilter {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113
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
