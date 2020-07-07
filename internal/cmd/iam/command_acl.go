package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/confluentinc/cli/internal/pkg/examples"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewACLCommand returns the Cobra command for ACLs.
func NewACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &aclCommand{
		AuthenticatedCLICommand: pcmd.NewAuthenticatedWithMDSCLICommand(&cobra.Command{
			Use:   "acl",
			Short: `Manage Kafka ACLs (5.4+ only).`,
		}, prerunner),
	}
	cmd.init()
	return cmd.Command
}

func (c *aclCommand) init() {
	cmd := &cobra.Command{
		Use:   "create",
		Args:  cobra.NoArgs,
		Short: "Create a Kafka ACL.",
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Create an ACL that grants the specified user READ permission to the specified consumer group in the specified Kafka cluster:",
				Code: "confluent iam acl create --allow --principal User:User1 --operation READ --consumer-group java_example_group_1 --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Desc: "Create an ACL that grants the specified user WRITE permission on all topics in the specified Kafka cluster.",
				Code: "confluent iam acl create --allow --principal User:User1 --operation WRITE --topic '*' --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Desc: "Create an ACL that assigns a group READ access to all topics that use the specified prefix in the specified Kafka cluster.",
				Code: "confluent iam acl create --allow --principal Group:Finance --operation READ --topic financial --prefix --kafka-cluster-id <kafka-cluster-id>",
			},
		),
	}
	cmd.Flags().AddFlagSet(addAclFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "delete",
		Short: `Delete a Kafka ACL.`,
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "Delete an ACL that granted the specified user access to the Test topic in the specified cluster:",
				Code: "confluent iam acl delete --kafka-cluster-id <kafka-cluster-id> --allow --principal User:Jane --topic Test",
			},
		),
		RunE: c.delete,
		Args: cobra.NoArgs,
	}
	cmd.Flags().AddFlagSet(deleteAclFlags())
	cmd.Flags().SortFlags = false

	c.AddCommand(cmd)

	cmd = &cobra.Command{
		Use:   "list",
		Short: "List Kafka ACLs for a resource.",
		Example: examples.BuildExampleString(
			examples.Example{
				Desc: "List all the ACLs for the specified Kafka cluster:",
				Code: "confluent iam acl list --kafka-cluster-id <kafka-cluster-id>",
			},
			examples.Example{
				Desc: "List all the ACLs for the specified cluster that include allow permissions for the user Jane:",
				Code: "confluent iam acl list --kafka-cluster-id <kafka-cluster-id> --allow --principal User:Jane",
			},
		),
		RunE: c.list,
		Args: cobra.NoArgs,
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
		aclConfiguration.AclBinding.Pattern.PatternType = mds.PATTERNTYPE_LITERAL
	}

	if aclConfiguration.AclBinding.Pattern.ResourceType == "" {
		aclConfiguration.errors = multierror.Append(aclConfiguration.errors, fmt.Errorf("exactly one of %v must be set",
			convertToFlags(mds.ACLRESOURCETYPE_TOPIC, mds.ACLRESOURCETYPE_GROUP,
				mds.ACLRESOURCETYPE_CLUSTER, mds.ACLRESOURCETYPE_TRANSACTIONAL_ID)))
	}
	return aclConfiguration
}

// convertToFilter converts a CreateAclRequest to an ACLFilterRequest
func convertToAclFilterRequest(request *mds.CreateAclRequest) mds.AclFilterRequest {
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
