package kafka

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
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

func newAclCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage Kafka ACLs.",
	}

	c := &aclCommand{pcmd.NewAuthenticatedStateFlagCommand(cmd, prerunner)}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newCreateCommand())
		cmd.AddCommand(c.newDeleteCommand())
		cmd.AddCommand(c.newListCommand())
	} else {
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)
		cmd.AddCommand(c.newCreateCommandOnPrem())
		cmd.AddCommand(c.newDeleteCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}

// validateAddAndDelete ensures the minimum requirements for acl add and delete are met
func validateAddAndDelete(binding *ACLConfiguration) {
	if binding.Entry.Principal == "" {
		err := fmt.Errorf(errors.ExactlyOneSetErrorMsg, "service-account, principal")
		binding.errors = multierror.Append(binding.errors, err)
	}

	if binding.Entry.PermissionType == schedv1.ACLPermissionTypes_UNKNOWN {
		err := fmt.Errorf(errors.MustSetAllowOrDenyErrorMsg)
		binding.errors = multierror.Append(binding.errors, err)
	}

	if binding.Pattern.PatternType == schedv1.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = schedv1.PatternTypes_LITERAL
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == schedv1.ResourceTypes_UNKNOWN {
		err := fmt.Errorf(errors.MustSetResourceTypeErrorMsg,
			listEnum(schedv1.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"}))
		binding.errors = multierror.Append(binding.errors, err)
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
		principal := acl[i].ACLBinding.Entry.Principal
		if principal != "" {
			resourceId, err := parsePrincipal(principal)
			if err != nil {
				return errors.Wrap(err, "failed to parse principal")
			}
			if resource.LookupType(resourceId) == resource.User || resource.LookupType(resourceId) == resource.ServiceAccount {
				userId, ok := idMap[resourceId]
				if !ok {
					return fmt.Errorf(errors.PrincipalNotFoundErrorMsg, resourceId)
				}
				resourceId = strconv.Itoa(int(userId))
			}
			acl[i].ACLBinding.Entry.Principal = fmt.Sprintf("User:%s", resourceId)
		}
	}
	return nil
}

func parsePrincipal(principal string) (string, error) {
	if !strings.HasPrefix(principal, "User:") {
		return "", fmt.Errorf(errors.BadPrincipalErrorMsg)
	}
	resourceId := strings.SplitN(principal, ":", 2)[1]
	return resourceId, nil
}

func (c *aclCommand) mapUserIdToResourceId() (map[int32]string, error) {
	serviceAccounts, err := c.PrivateClient.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	adminUsers, err := c.PrivateClient.User.List(context.Background())
	if err != nil {
		return nil, err
	}

	users := append(serviceAccounts, adminUsers...)

	idMap := make(map[int32]string)
	for _, sa := range users {
		idMap[sa.Id] = sa.ResourceId
	}
	return idMap, nil
}

func (c *aclCommand) mapResourceIdToUserId() (map[string]int32, error) {
	serviceAccounts, err := c.PrivateClient.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	adminUsers, err := c.PrivateClient.User.List(context.Background())
	if err != nil {
		return nil, err
	}

	users := append(serviceAccounts, adminUsers...)

	idMap := make(map[string]int32)
	for _, sa := range users {
		idMap[sa.ResourceId] = sa.Id
	}
	return idMap, nil
}

func (c *aclCommand) provisioningClusterCheck(lkc string) error {
	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, c.EnvironmentId())
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}
	if cluster.Status.Phase == ccloudv2.StatusProvisioning {
		return errors.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
	}
	return nil
}
