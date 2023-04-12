package kafka

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

type aclCommand struct {
	*pcmd.AuthenticatedCLICommand
}

func newAclCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "acl",
		Short: "Manage Kafka ACLs.",
	}

	c := &aclCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

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

	if binding.Entry.PermissionType == ccstructs.ACLPermissionTypes_UNKNOWN {
		err := fmt.Errorf(errors.MustSetAllowOrDenyErrorMsg)
		binding.errors = multierror.Append(binding.errors, err)
	}

	if binding.Pattern.PatternType == ccstructs.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = ccstructs.PatternTypes_LITERAL
	}

	if binding.Pattern == nil || binding.Pattern.ResourceType == ccstructs.ResourceTypes_UNKNOWN {
		err := fmt.Errorf(errors.MustSetResourceTypeErrorMsg,
			listEnum(ccstructs.ResourceTypes_ResourceType_name, []string{"ANY", "UNKNOWN"}))
		binding.errors = multierror.Append(binding.errors, err)
	}
}

// convertToFilter converts a ACLBinding to a KafkaAPIACLFilterRequest
func convertToFilter(binding *ccstructs.ACLBinding) *ccstructs.ACLFilter {
	// ACE matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/acl/AccessControlEntryFilter.java#L102-L113
	if binding.Entry == nil {
		binding.Entry = new(ccstructs.AccessControlEntryConfig)
	}

	if binding.Entry.Operation == ccstructs.ACLOperations_UNKNOWN {
		binding.Entry.Operation = ccstructs.ACLOperations_ANY
	}

	if binding.Entry.PermissionType == ccstructs.ACLPermissionTypes_UNKNOWN {
		binding.Entry.PermissionType = ccstructs.ACLPermissionTypes_ANY
	}

	// ResourcePattern matching rules
	// https://github.com/apache/kafka/blob/trunk/clients/src/main/java/org/apache/kafka/common/resource/ResourcePatternFilter.java#L42-L56
	if binding.Pattern == nil {
		binding.Pattern = &ccstructs.ResourcePatternConfig{}
	}

	binding.Entry.Host = "*"

	if binding.Pattern.ResourceType == ccstructs.ResourceTypes_UNKNOWN {
		binding.Pattern.ResourceType = ccstructs.ResourceTypes_ANY
	}

	if binding.Pattern.PatternType == ccstructs.PatternTypes_UNKNOWN {
		binding.Pattern.PatternType = ccstructs.PatternTypes_ANY
	}

	return &ccstructs.ACLFilter{
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
		return "", fmt.Errorf(`principal must begin with "User:"`)
	}

	id := strings.TrimPrefix(principal, "User:")

	if _, err := strconv.Atoi(id); err == nil {
		return "", fmt.Errorf("numeric IDs are not supported")
	}

	return id, nil
}

func (c *aclCommand) getAllUsers() ([]*ccloud.User, error) {
	serviceAccounts, err := c.Client.User.GetServiceAccounts(context.Background())
	if err != nil {
		return nil, err
	}

	adminUsers, err := c.Client.User.List(context.Background())
	if err != nil {
		return nil, err
	}

	return append(serviceAccounts, adminUsers...), nil
}

func (c *aclCommand) mapUserIdToResourceId(users []*ccloud.User) map[int32]string {
	idMap := make(map[int32]string)
	for _, sa := range users {
		idMap[sa.Id] = sa.ResourceId
	}
	return idMap
}

func (c *aclCommand) mapResourceIdToUserId(users []*ccloud.User) map[string]int32 {
	idMap := make(map[string]int32)
	for _, sa := range users {
		idMap[sa.ResourceId] = sa.Id
	}
	return idMap
}

func (c *aclCommand) provisioningClusterCheck(lkc string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}
	cluster, httpResp, err := c.V2Client.DescribeKafkaCluster(lkc, environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, lkc, httpResp)
	}
	if cluster.Status.Phase == ccloudv2.StatusProvisioning {
		return errors.Errorf(errors.KafkaRestProvisioningErrorMsg, lkc)
	}
	return nil
}
