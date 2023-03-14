package iam

import (
	"fmt"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
)

const rbacPromptMsg = "Are you sure you want to delete this role binding?"

func (c *roleBindingCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a role binding.",
		Args:  cobra.NoArgs,
		RunE:  c.delete,
	}

	if c.cfg.IsCloudLogin() {
		cmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: `Delete the role "ResourceOwner" for the resource "Topic:my-topic" on the Kafka cluster "lkc-123456":`,
				Code: "confluent iam rbac role-binding delete --principal User:u-123456 --role ResourceOwner --environment env-12345 --kafka-cluster lkc-123456 --resource Topic:my-topic",
			},
		)
	}

	cmd.Flags().String("role", "", "Role name of the existing role binding.")
	cmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	pcmd.AddForceFlag(cmd)
	addClusterFlags(cmd, c.cfg.IsCloudLogin(), c.CLICommand)
	cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func (c *roleBindingCommand) delete(cmd *cobra.Command, _ []string) error {
	isCloud := c.cfg.IsCloudLogin()

	if isCloud {
		deleteRoleBinding, err := c.parseV2RoleBinding(cmd)
		if err != nil {
			return err
		}

		err = c.ccloudDelete(cmd, deleteRoleBinding)
		if err != nil {
			return err
		}

		return c.displayCCloudCreateAndDeleteOutput(cmd, deleteRoleBinding)
	} else {
		options, err := c.parseCommon(cmd)
		if err != nil {
			return err
		}
		httpResp, err := c.confluentDelete(cmd, options)
		if err != nil {
			return err
		}

		if httpResp != nil && httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, httpResp.StatusCode), errors.HTTPStatusCodeSuggestions)
		}

		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudDelete(cmd *cobra.Command, deleteRoleBinding *mdsv2.IamV2RoleBinding) error {
	roleBindings, err := c.V2Client.ListIamRoleBindings(deleteRoleBinding.GetCrnPattern(), deleteRoleBinding.GetPrincipal(), deleteRoleBinding.GetRoleName())
	if err != nil {
		return err
	}

	var roleBindingToDelete *mdsv2.IamV2RoleBinding
	for _, roleBinding := range roleBindings {
		if roleBinding.GetCrnPattern() == deleteRoleBinding.GetCrnPattern() {
			roleBindingToDelete = &roleBinding
			break
		}
	}

	if roleBindingToDelete == nil {
		return errors.NewErrorWithSuggestions(errors.RoleBindingNotFoundErrorMsg, errors.RoleBindingNotFoundSuggestions)
	}

	if ok, err := form.ConfirmDeletionYesNoCustomPrompt(cmd, rbacPromptMsg); err != nil || !ok {
		return err
	}

	_, err = c.V2Client.DeleteIamRoleBinding(roleBindingToDelete.GetId())
	return err
}

func (c *roleBindingCommand) confluentDelete(cmd *cobra.Command, options *roleBindingOptions) (*http.Response, error) {
	if ok, err := form.ConfirmDeletionYesNoCustomPrompt(cmd, rbacPromptMsg); err != nil || !ok {
		return nil, err
	}

	if options.resource != "" {
		return c.MDSClient.RBACRoleBindingCRUDApi.RemoveRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else {
		return c.MDSClient.RBACRoleBindingCRUDApi.DeleteRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.mdsScope)
	}
}
