package iam

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"os"

	acl_util "github.com/confluentinc/cli/internal/pkg/acl"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/mds-sdk-go"
)

type aclCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
	ctx    context.Context
}

// NewACLCommand returns the Cobra command for ACLs.
func NewACLCommand(config *config.Config, ch *pcmd.ConfigHelper, client *mds.APIClient) *cobra.Command {
	cmd := &aclCommand{
		Command: &cobra.Command{
			Use:   "acl",
			Short: `Manage Kafka ACLs.`,
		},
		config: config,
		client: client,
		ctx:    context.WithValue(context.Background(), mds.ContextAccessToken, config.AuthToken),
	}

	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	cliName := c.config.CLIName

	cmd := &cobra.Command{
		Use:   "create",
		Short: `Create a Kafka ACL.`,
		Example: `You can only specify one of these flags per command invocation: ` + "``cluster``, ``consumer-group``" + `,
` + "``topic``, or ``transactional-id``" + ` per command invocation. For example, if you want to specify both
` + "``consumer-group`` and ``topic``" + `, you must specify this as two separate commands:

::

	` + cliName + ` iam acl create --allow --principal User:1522 --operation READ --consumer-group \
	java_example_group_1 --kafka-cluster-id my-cluster

::

	` + cliName + ` iam acl create --allow --principal User:1522 --operation READ --topic '*' \
	--kafka-cluster-id my-cluster

`,
		RunE: c.create,
		Args: cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(addAclFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: `Delete a Kafka ACL.`,
		RunE:  c.delete,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(deleteAclFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: `List Kafka ACLs for a resource.`,
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(listAclFlags())
	cmd.Flags().Int("service-account-id", 0, "Service account ID.")
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)
}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl := parseMds(cmd)

	bindings, _, err := c.client.KafkaACLManagementApi.SearchAclBinding(c.ctx, convertToAclFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	acl_util.PrintMdsAcls(bindings, os.Stdout)
	return nil
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAclAdd(parseMds(cmd))

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}

	_, err := c.client.KafkaACLManagementApi.AddAclBinding(c.ctx, *acl.CreateAclRequest)

	return errors.HandleCommon(err, cmd)
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := parseMds(cmd)

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}

	bindings, _, err := c.client.KafkaACLManagementApi.RemoveAclBindings(c.ctx, convertToAclFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	acl_util.PrintMdsAcls(bindings, os.Stdout)
	return nil
}

// validateAclAdd ensures the minimum requirements for acl add is met
func validateAclAdd(aclConfiguration *ACLConfiguration) *ACLConfiguration {
	if aclConfiguration.AclBinding.Entry.PermissionType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, fmt.Errorf("--allow or --deny must be set when adding or deleting an ACL"))
	}

	if aclConfiguration.AclBinding.Pattern.PatternType == "" {
		aclConfiguration.AclBinding.Pattern.PatternType = mds.PATTERN_TYPE_LITERAL
	}

	if aclConfiguration.AclBinding.Pattern.ResourceType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(mds.ACL_RESOURCE_TYPE_TOPIC, mds.ACL_RESOURCE_TYPE_GROUP,
				mds.ACL_RESOURCE_TYPE_CLUSTER, mds.ACL_RESOURCE_TYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

// convertToFilter converts a CreateAclRequest to an ACLFilterRequest
func convertToAclFilterRequest(request *mds.CreateAclRequest) mds.AclFilterRequest {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113

	if request.AclBinding.Entry.Operation == "" {
		request.AclBinding.Entry.Operation = mds.ACL_OPERATION_ANY
	}

	if request.AclBinding.Entry.PermissionType == "" {
		request.AclBinding.Entry.PermissionType = mds.ACL_PERMISSION_TYPE_ANY
	}
	// delete/list shouldn't provide a host value
	request.AclBinding.Entry.Host = ""

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if request.AclBinding.Pattern.ResourceType == "" {
		request.AclBinding.Pattern.ResourceType = mds.ACL_RESOURCE_TYPE_ANY
	}

	if request.AclBinding.Pattern.PatternType == "" {
		if request.AclBinding.Pattern.Name == "" {
			request.AclBinding.Pattern.PatternType = mds.PATTERN_TYPE_ANY
		} else {
			request.AclBinding.Pattern.PatternType = mds.PATTERN_TYPE_LITERAL
		}
	}

	return mds.AclFilterRequest {
		Scope: request.Scope,
		AclBindingFilter: mds.AclBindingFilter {
			EntryFilter:   mds.AccessControlEntryFilter{
				Host:           request.AclBinding.Entry.Host,
				Operation:      request.AclBinding.Entry.Operation,
				PermissionType: request.AclBinding.Entry.PermissionType,
				Principal:      request.AclBinding.Entry.Principal,
			},
			PatternFilter: mds.KafkaResourcePatternFilter{
				ResourceType: request.AclBinding.Pattern.ResourceType,
				Name:         request.AclBinding.Pattern.Name,
				PatternType:  request.AclBinding.Pattern.PatternType,
			},
		},
	}
}
