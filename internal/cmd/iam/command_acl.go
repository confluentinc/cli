package iam

import (
	"context"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"net/http"

	"github.com/confluentinc/mds-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewACLCommand returns the Cobra command for ACLs.
func NewACLCommand(config *v3.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &aclCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(&cobra.Command{
			Use:   "acl",
			Short: `Manage Kafka ACLs (5.4+ only).`,
		}, config, prerunner),
	}
	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	cliName := c.Config.CLIName

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
	cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)
}

func (c *aclCommand) list(cmd *cobra.Command, args []string) error {
	acl := parse(cmd)

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.SearchAclBinding(c.createContext(), convertToAclFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return c.handleAclError(cmd, err, response)
	}
	return PrintAcls(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}

func (c *aclCommand) create(cmd *cobra.Command, args []string) error {
	acl := validateAclAddDelete(parse(cmd))

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}

	response, err := c.MDSClient.KafkaACLManagementApi.AddAclBinding(c.createContext(), *acl.CreateAclRequest)

	if err != nil {
		return c.handleAclError(cmd, err, response)
	}

	return nil
}

func (c *aclCommand) delete(cmd *cobra.Command, args []string) error {
	acl := parse(cmd)

	if acl.errors != nil {
		return errors.HandleCommon(acl.errors, cmd)
	}

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.RemoveAclBindings(c.createContext(), convertToAclFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return c.handleAclError(cmd, err, response)
	}

	return PrintAcls(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}

func (c *aclCommand) handleAclError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		cmd.SilenceUsage = true
		return fmt.Errorf("Unable to %s ACLs (%s). Ensure that you're running against MDS with CP 5.4+.", cmd.Name(), err.Error())
	}
	return errors.HandleCommon(err, cmd)
}

// validateAclAddDelete ensures the minimum requirements for acl add/delete is met
func validateAclAddDelete(aclConfiguration *ACLConfiguration) *ACLConfiguration {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
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

	return mds.AclFilterRequest{
		Scope: request.Scope,
		AclBindingFilter: mds.AclBindingFilter{
			EntryFilter: mds.AccessControlEntryFilter{
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

func PrintAcls(cmd *cobra.Command, kafkaClusterId string, bindingsObj []mds.AclBinding) error {
	var fields = []string{"KafkaClusterId", "Principal", "Permission", "Operation", "Host", "Resource", "Name", "Type"}
	var structuredRenames = []string{"kafka_cluster_id", "principal", "permission", "operation", "host", "resource", "name", "type"}

	// delete also uses this function but doesn't have -o flag defined, -o flag is needed NewListOutputWriter
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		cmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	}

	outputWriter, err := output.NewListOutputWriter(cmd, fields, fields, structuredRenames)
	if err != nil {
		return err
	}
	for _, binding := range bindingsObj {

		record := &struct {
			KafkaClusterId string
			Principal      string
			Permission     mds.AclPermissionType
			Operation      mds.AclOperation
			Host           string
			Resource       mds.AclResourceType
			Name           string
			Type           mds.PatternType
		}{
			kafkaClusterId,
			binding.Entry.Principal,
			binding.Entry.PermissionType,
			binding.Entry.Operation,
			binding.Entry.Host,
			binding.Pattern.ResourceType,
			binding.Pattern.Name,
			binding.Pattern.PatternType,
		}
		outputWriter.AddElement(record)
	}
	return outputWriter.Out()
}

func (c *aclCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.AuthToken())
}
