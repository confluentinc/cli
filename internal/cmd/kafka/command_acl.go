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
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewACLCommand returns the Cobra command for Kafka ACL.
func NewACLCommand(prerunner pcmd.PreRunner, config *v2.Config) *cobra.Command {
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

	cmd := &cobra.Command{
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
	cmd.Flags().AddFlagSet(aclConfigFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: `Delete a Kafka ACL.`,
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(aclConfigFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: `List Kafka ACLs for a resource.`,
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(resourceFlags())
	cmd.Flags().Int("service-account", 0, "Service account ID.")
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)
}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl := parse(cmd)

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	resp, err := c.Client.Kafka.ListACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return aclutil.PrintAcls(cmd, resp, os.Stdout)
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}
	err = c.Client.Kafka.CreateACL(context.Background(), cluster, []*kafkav1.ACLBinding{acl.ACLBinding})

	return errors.HandleCommon(err, cmd)
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := validateAddDelete(parse(cmd))

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}
	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = c.Client.Kafka.DeleteACL(context.Background(), cluster, convertToFilter(acl.ACLBinding))

	return errors.HandleCommon(err, cmd)
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(binding *ACLConfiguration) *ACLConfiguration {
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
