package iam

import (
	"context"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type out struct {
	KafkaClusterId string                  `human:"Kafka Cluster" serialized:"kafka_cluster_id"`
	Principal      string                  `human:"Principal" serialized:"principal"`
	Permission     mdsv1.AclPermissionType `human:"Permission" serialized:"permission"`
	Operation      mdsv1.AclOperation      `human:"Operation" serialized:"operation"`
	Host           string                  `human:"Host" serialized:"host"`
	ResourceType   mdsv1.AclResourceType   `human:"Resource Type" serialized:"resource_type"`
	ResourceName   string                  `human:"Resource Name" serialized:"resource_name"`
	PatternType    mdsv1.PatternType       `human:"Pattern Type" serialized:"pattern_type"`
}

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newAclCommand(prerunner pcmd.PreRunner) *cobra.Command {
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

func (c *aclCommand) handleAclError(cmd *cobra.Command, err error, response *http.Response) error {
	if response != nil && response.StatusCode == http.StatusNotFound {
		return fmt.Errorf("unable to %s ACLs: %w", cmd.Name(), err)
	}
	return err
}

// convertToFilter converts a CreateAclRequest to an AclFilterRequest
func convertToAclFilterRequest(request *mdsv1.CreateAclRequest) mdsv1.AclFilterRequest {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113

	if request.AclBinding.Entry.Operation == "" {
		request.AclBinding.Entry.Operation = mdsv1.ACLOPERATION_ANY
	}

	if request.AclBinding.Entry.PermissionType == "" {
		request.AclBinding.Entry.PermissionType = mdsv1.ACLPERMISSIONTYPE_ANY
	}
	// delete/list shouldn't provide a host value
	request.AclBinding.Entry.Host = ""

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if request.AclBinding.Pattern.ResourceType == "" {
		request.AclBinding.Pattern.ResourceType = mdsv1.ACLRESOURCETYPE_ANY
	}

	if request.AclBinding.Pattern.PatternType == "" {
		if request.AclBinding.Pattern.Name == "" {
			request.AclBinding.Pattern.PatternType = mdsv1.PATTERNTYPE_ANY
		} else {
			request.AclBinding.Pattern.PatternType = mdsv1.PATTERNTYPE_LITERAL
		}
	}

	return mdsv1.AclFilterRequest{
		Scope: request.Scope,
		AclBindingFilter: mdsv1.AclBindingFilter{
			EntryFilter: mdsv1.AccessControlEntryFilter{
				Host:           request.AclBinding.Entry.Host,
				Operation:      request.AclBinding.Entry.Operation,
				PermissionType: request.AclBinding.Entry.PermissionType,
				Principal:      request.AclBinding.Entry.Principal,
			},
			PatternFilter: mdsv1.KafkaResourcePatternFilter{
				ResourceType: request.AclBinding.Pattern.ResourceType,
				Name:         request.AclBinding.Pattern.Name,
				PatternType:  request.AclBinding.Pattern.PatternType,
			},
		},
	}
}

func printAcls(cmd *cobra.Command, kafkaClusterId string, aclBindings []mdsv1.AclBinding) error {
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
	return context.WithValue(context.Background(), mdsv1.ContextAccessToken, c.Context.GetAuthToken())
}
