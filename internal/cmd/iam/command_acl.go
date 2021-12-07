package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type aclCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func NewACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "acl",
		Short:       "Manage Kafka ACLs (5.4+ only).",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &aclCommand{pcmd.NewAuthenticatedWithMDSStateFlagCommand(cmd, prerunner, AclSubcommandFlags)}

	c.AddCommand(c.newCreateCommand())
	c.AddCommand(c.newDeleteCommand())
	c.AddCommand(c.newListCommand())

	return c.Command
}

func (c *aclCommand) handleACLError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.UnableToPerformAclErrorMsg, cmd.Name(), err.Error()), errors.UnableToPerformAclSuggestions)
	}
	return err
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

func printACLs(cmd *cobra.Command, kafkaClusterId string, bindingsObj []mds.AclBinding) error {
	var fields = []string{"KafkaClusterId", "Principal", "Permission", "Operation", "Host", "ResourceType", "ResourceName", "PatternType"}
	var structuredRenames = []string{"kafka_cluster_id", "principal", "permission", "operation", "host", "resource_type", "resource_name", "pattern_type"}

	// delete also uses this function but doesn't have -o flag defined, -o flag is needed for NewListOutputWriter initializers
	_, err := cmd.Flags().GetString(output.FlagName)
	if err != nil {
		output.AddFlag(cmd)
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
