package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type out struct {
	KafkaClusterId string                `human:"Kafka Cluster" serialized:"kafka_cluster_id"`
	Principal      string                `human:"Principal" serialized:"principal"`
	Permission     mds.AclPermissionType `human:"Permission" serialized:"permission"`
	Operation      mds.AclOperation      `human:"Operation" serialized:"operation"`
	Host           string                `human:"Host" serialized:"host"`
	ResourceType   mds.AclResourceType   `human:"Resource Type" serialized:"resource_type"`
	ResourceName   string                `human:"Resource Name" serialized:"resource_name"`
	PatternType    mds.PatternType       `human:"Pattern Type" serialized:"pattern_type"`
}

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newACLCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "acl",
		Short:       "Manage centralized ACLs.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
	}

	c := &aclCommand{pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
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

func printACLs(cmd *cobra.Command, kafkaClusterId string, aclBindings []mds.AclBinding) error {
	list := output.NewList(cmd)
	for _, binding := range aclBindings {
		list.Add(&out{
			KafkaClusterId: kafkaClusterId,
			Principal:      binding.Entry.Principal,
			Permission:     binding.Entry.PermissionType,
			Operation:      binding.Entry.Operation,
			Host:           binding.Entry.Host,
			ResourceType:   binding.Pattern.ResourceType,
			ResourceName:   binding.Pattern.Name,
			PatternType:    binding.Pattern.PatternType,
		})
	}
	return list.Print()
}

func (c *aclCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.Context.GetAuthToken())
}
