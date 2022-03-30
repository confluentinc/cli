package kafka

import (
	"context"
	"fmt"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

var (
	listFieldsOnPrem       = []string{"Principal", "Permission", "Operation", "Host", "ResourceType", "ResourceName", "PatternType"}
	humanLabelsOnPrem      = []string{"Principal", "Permission", "Operation", "Host", "Resource Type", "Resource Name", "Pattern Type"}
	structuredLabelsOnPrem = []string{"principal", "permission", "operation", "host", "resource_type", "resource_name", "pattern_type"}
)

type aclCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
}

func newAclCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *aclCommand {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage Kafka ACLs.",
	}

	c := &aclCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	if cfg.IsCloudLogin() {
		c.AddCommand(c.newCreateCommand())
		c.AddCommand(c.newDeleteCommand())
		c.AddCommand(c.newListCommand())
	} else {
		c.SetPersistentPreRunE(prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand))

		c.AddCommand(c.newCreateCommandOnPrem())
		c.AddCommand(c.newDeleteCommandOnPrem())
		c.AddCommand(c.newListCommandOnPrem())
	}

	return c
}

// validateAddAndDelete ensures the minimum requirements for acl add and delete are met
func validateAddAndDelete(binding *ACLConfiguration) {
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

func (c *aclCommand) aclResourceIdToNumericId(acl []*ACLConfiguration, idMap map[string]int32) error {
	for i := 0; i < len(acl); i++ {
		if acl[i].ACLBinding.Entry.Principal != "" { // it has a service-account flag
			serviceAccountID := acl[i].ACLBinding.Entry.Principal[5:] // extract service account id
			if resource.LookupType(serviceAccountID) != resource.ServiceAccount {
				return errors.New(errors.BadServiceAccountIDErrorMsg)
			}
			userId, ok := idMap[serviceAccountID]
			if !ok {
				return fmt.Errorf(errors.ServiceAccountNotFoundErrorMsg, serviceAccountID)
			}
			acl[i].ACLBinding.Entry.Principal = fmt.Sprintf("User:%d", userId) // translate into numeric ID
		}
	}
	return nil
}

func (c *aclCommand) mapUserIdToResourceId() (map[int32]string, error) {
	serviceAccounts, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	idMap := make(map[int32]string)
	for _, sa := range serviceAccounts {
		idMap[sa.Id] = sa.ResourceId
	}
	return idMap, nil
}

func (c *aclCommand) mapResourceIdToUserId() (map[string]int32, error) {
	serviceAccounts, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	idMap := make(map[string]int32)
	for _, sa := range serviceAccounts {
		idMap[sa.ResourceId] = sa.Id
	}
	return idMap, nil
}
