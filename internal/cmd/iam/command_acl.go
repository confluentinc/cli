package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type aclCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

// NewACLCommand returns the Cobra command for ACLs.
func NewACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &aclCommand{
		AuthenticatedStateFlagCommand: pcmd.NewAuthenticatedWithMDSStateFlagCommand(&cobra.Command{
			Use:         "acl",
			Short:       "Manage Kafka ACLs (5.4+ only).",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		}, prerunner, AclSubcommandFlags),
	}

	cmd.init()

	return cmd.Command
}

func (c *aclCommand) init() {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Kafka ACL.",
		Long:  "Create a Kafka ACL.\n\nThis command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an ACL that grants the specified user READ permission to the specified consumer group in the specified Kafka cluster:",
				Code: version.CLIName + " iam acl create --allow --principal User:User1 --operation READ --consumer-group java_example_group_1 --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Text: "Create an ACL that grants the specified user WRITE permission on all topics in the specified Kafka cluster.",
				Code: version.CLIName + " iam acl create --allow --principal User:User1 --operation WRITE --topic '*' --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Text: "Create an ACL that assigns a group READ access to all topics that use the specified prefix in the specified Kafka cluster.",
				Code: version.CLIName + " iam acl create --allow --principal Group:Finance --operation READ --topic financial --prefix --kafka-cluster-id <kafka-cluster-id>",
			},
		),
	}
	cmd.Flags().AddFlagSet(addACLFlags())

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete a Kafka ACL.",
		Long:  "Delete a Kafka ACL.\n\nThis command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Delete an ACL that granted the specified user access to the Test topic in the specified cluster:",
				Code: version.CLIName + " iam acl delete --kafka-cluster-id <kafka-cluster-id> --allow --principal User:Jane --topic Test",
			},
		),
	}
	cmd.Flags().AddFlagSet(deleteACLFlags())

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Long:  "List Kafka ACLs for a resource.\n\nThis command only works with centralized ACLs.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List all the ACLs for the specified Kafka cluster:",
				Code: version.CLIName + " iam acl list --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Text: "List all the ACLs for the specified cluster that include allow permissions for the user Jane:",
				Code: version.CLIName + " iam acl list --kafka-cluster-id <kafka-cluster-id> --allow --principal User:Jane",
			},
		),
	}
	cmd.Flags().AddFlagSet(listACLFlags())
	pcmd.AddOutputFlag(cmd)

	c.AddCommand(cmd)
}

func (c *aclCommand) list(cmd *cobra.Command, _ []string) error {
	acl := parse(cmd)

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.SearchAclBinding(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return c.handleACLError(cmd, err, response)
	}
	return PrintACLs(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}

func (c *aclCommand) create(cmd *cobra.Command, _ []string) error {
	acl := validateACLAddDelete(parse(cmd))

	if acl.errors != nil {
		return acl.errors
	}

	response, err := c.MDSClient.KafkaACLManagementApi.AddAclBinding(c.createContext(), *acl.CreateAclRequest)

	if err != nil {
		return c.handleACLError(cmd, err, response)
	}
	return PrintACLs(cmd, acl.CreateAclRequest.Scope.Clusters.KafkaCluster, []mds.AclBinding{acl.CreateAclRequest.AclBinding})
}

func (c *aclCommand) delete(cmd *cobra.Command, _ []string) error {
	acl := parse(cmd)

	if acl.errors != nil {
		return acl.errors
	}

	bindings, response, err := c.MDSClient.KafkaACLManagementApi.RemoveAclBindings(c.createContext(), convertToACLFilterRequest(acl.CreateAclRequest))

	if err != nil {
		return c.handleACLError(cmd, err, response)
	}

	return PrintACLs(cmd, acl.Scope.Clusters.KafkaCluster, bindings)
}

func (c *aclCommand) handleACLError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnableToPerformAclErrorMsg, cmd.Name(), err.Error()), errors.UnableToPerformAclSuggestions)
	}
	return err
}

// validateACLAddDelete ensures the minimum requirements for acl add/delete is met
func validateACLAddDelete(aclConfiguration *ACLConfiguration) *ACLConfiguration {
	// delete is deliberately less powerful in the cli than in the API to prevent accidental
	// deletion of too many acls at once. Expectation is that multi delete will be done via
	// repeated invocation of the cli by external scripts.
	if aclConfiguration.AclBinding.Entry.PermissionType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetAllowOrDenyErrorMsg))
	}

	if aclConfiguration.AclBinding.Pattern.PatternType == "" {
		aclConfiguration.AclBinding.Pattern.PatternType = mds.PATTERNTYPE_LITERAL
	}

	if aclConfiguration.AclBinding.Pattern.ResourceType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, errors.Errorf(errors.MustSetResourceTypeErrorMsg,
			convertToFlags(mds.ACLRESOURCETYPE_TOPIC, mds.ACLRESOURCETYPE_GROUP,
				mds.ACLRESOURCETYPE_CLUSTER, mds.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

// convertToFilter converts a CreateAclRequest to an AclFilterRequest
func convertToACLFilterRequest(request *mds.CreateAclRequest) mds.AclFilterRequest {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113

	if request.AclBinding.Entry.Operation == "" {
		request.AclBinding.Entry.Operation = mds.ACLOPERATION_ANY
	}

	if request.AclBinding.Entry.PermissionType == "" {
		request.AclBinding.Entry.PermissionType = mds.ACLPERMISSIONTYPE_ANY
	}
	// delete/list shouldn't provide a host value
	request.AclBinding.Entry.Host = ""

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if request.AclBinding.Pattern.ResourceType == "" {
		request.AclBinding.Pattern.ResourceType = mds.ACLRESOURCETYPE_ANY
	}

	if request.AclBinding.Pattern.PatternType == "" {
		if request.AclBinding.Pattern.Name == "" {
			request.AclBinding.Pattern.PatternType = mds.PATTERNTYPE_ANY
		} else {
			request.AclBinding.Pattern.PatternType = mds.PATTERNTYPE_LITERAL
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

func PrintACLs(cmd *cobra.Command, kafkaClusterId string, bindingsObj []mds.AclBinding) error {
	var fields = []string{"KafkaClusterId", "Principal", "Permission", "Operation", "Host", "ResourceType", "ResourceName", "PatternType"}
	var structuredRenames = []string{"kafka_cluster_id", "principal", "permission", "operation", "host", "resource_type", "resource_name", "pattern_type"}

	// delete also uses this function but doesn't have -o flag defined, -o flag is needed for NewListOutputWriter initializers
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		pcmd.AddOutputFlag(cmd)
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
			ResourceType   mds.AclResourceType
			ResourceName   string
			PatternType    mds.PatternType
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
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
}
