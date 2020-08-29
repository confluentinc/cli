package kafka

import (
	"context"
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/examples"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"

	aclutil "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
func NewACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "acl",
			Short: "Manage Kafka ACLs.",
		}, prerunner)
	cmd := &aclCommand{AuthenticatedCLICommand: cliCmd}
	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	c.Command.PersistentFlags().String("cluster", "", "Kafka cluster ID.")

	createCmd = &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "You can only specify one of these flags per command invocation: ``cluster``, ``consumer-group``, ``topic``, or ``transactional-id``. For example, if you want to specify both ``consumer-group`` and ``topic``, you must specify this as two separate commands:",
				Code: "ccloud kafka acl create --allow --service-account 1522 --operation READ --consumer-group java_example_group_1\nccloud kafka acl create --allow --service-account 1522 --operation READ --topic '*'",
			},
		),
	}
	createCmd.Flags().AddFlagSet(aclConfigFlags())
	createCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	createCmd.Flags().SortFlags = false

	c.AddCommand(createCmd)

	deleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
	}
	deleteCmd.Flags().AddFlagSet(aclConfigFlags())
	deleteCmd.Flags().SortFlags = false

	c.AddCommand(deleteCmd)

	listCmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().AddFlagSet(resourceFlags())
	listCmd.Flags().Int("service-account", 0, "Service account ID.")
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false

	c.AddCommand(listCmd)
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl, err := parse(cmd)
	if err != nil {
		return err
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}
	resp, err := c.Client.Kafka.ListACLs(context.Background(), cluster, convertToFilter(acl[0].ACLBinding))

	if err != nil {
		return err
	}
	return aclutil.PrintACLs(cmd, resp, os.Stdout)
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}
	var bindings []*schedv1.ACLBinding
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		bindings = append(bindings, acl.ACLBinding)
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}

	err = c.Client.Kafka.CreateACLs(context.Background(), cluster, bindings)
	if err != nil {
		return err
	}
	return aclutil.PrintACLs(cmd, bindings, os.Stdout)
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acls, err := parse(cmd)
	if err != nil {
		return err
	}
	var filters []*schedv1.ACLFilter
	for _, acl := range acls {
		validateAddDelete(acl)
		if acl.errors != nil {
			return acl.errors
		}
		filters = append(filters, convertToFilter(acl.ACLBinding))
	}

	cluster, err := pcmd.KafkaCluster(cmd, c.Context)
	if err != nil {
		return err
	}
	err = c.Client.Kafka.DeleteACLs(context.Background(), cluster, filters)
	if err != nil {
		return err
	}
	pcmd.ErrPrintf(cmd, errors.DeletedACLsMsg)
	return nil
}

// validateAddDelete ensures the minimum requirements for acl add and delete are met
func validateAddDelete(binding *ACLConfiguration) {
	if binding.Entry.PermissionType == schedv1.ACLPermissionTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if binding.Pattern.PatternType == schedv1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = schedv1.PatternTypes_LITERAL
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == schedv1.ResourceTypes_UNKNOWN {
		binding.errors = multierror.Append(binding.errors, fmt.Errorf(errors.MustSetResourceTypeErrorMsg,
			listEnum(schedv1.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"})))
	}
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *schedv1.ACLBinding) *schedv1.ACLFilter {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113
	if binding.Entry == nil {
		binding.Entry = new(schedv1.AccessControlEntryConfig)
	}

	if binding.Entry.Operation == schedv1.ACLOperations_UNKNOWN {
		binding.Entry.Operation = schedv1.ACLOperations_ANY
	}

	if binding.Entry.PermissionType == schedv1.ACLPermissionTypes_UNKNOWN {
		binding.Entry.PermissionType = schedv1.ACLPermissionTypes_ANY
	}

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if binding.Pattern == nil {
		binding.Pattern = &schedv1.ResourcePatternConfig{}
	}

	binding.Entry.Host = "*"

	if binding.Pattern.ResourceType == schedv1.ResourceTypes_UNKNOWN {
		binding.Pattern.ResourceType = schedv1.ResourceTypes_ANY
	}

	if binding.Pattern.PatternType == schedv1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = schedv1.PatternTypes_ANY
	}

	return &schedv1.ACLFilter{
		EntryFilter:   binding.Entry,
		PatternFilter: binding.Pattern,
	}
}
