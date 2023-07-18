package kafka

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	nameconversions "github.com/confluentinc/cli/internal/pkg/name-conversions"
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

func (c *aclCommand) validateServiceAccountFlag(cmd *cobra.Command) error {
	if serviceAccount, err := cmd.Flags().GetString("service-account"); serviceAccount != "" {
		if err != nil {
			return err
		}
		serviceAccountId, err := nameconversions.ConvertIamServiceAccountNameToId(serviceAccount, c.V2Client, true)
		if err != nil {
			return nil
		}
		if err = cmd.Flags().Set("service-account", serviceAccountId); err != nil {
			return err
		}
	}
	return nil
}
