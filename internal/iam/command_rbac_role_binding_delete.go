package iam

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/spf13/cobra"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
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
				Code: "confluent iam rbac role-binding delete --principal User:u-123456 --role ResourceOwner --environment env-123456 --kafka-cluster lkc-123456 --resource Topic:my-topic",
			},
		)
	}

	cmd.Flags().String("role", "", "Role name of the existing role binding.")
	cmd.Flags().String("principal", "", `Principal type and identifier using "Prefix:ID" format.`)
	pcmd.AddForceFlag(cmd)
	addClusterFlags(cmd, c.cfg, c.CLICommand)
	cmd.Flags().String("resource", "", `Resource type and identifier using "Prefix:ID" format.`)
	cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("principal"))
	cobra.CheckErr(cmd.MarkFlagRequired("role"))

	return cmd
}

func (c *roleBindingCommand) delete(cmd *cobra.Command, _ []string) error {
	isCloud := c.cfg.IsCloudLogin()

	if isCloud {
		deleteRoleBinding, err := c.parseV2RoleBinding(cmd)
		if err != nil {
			return err
		}

		if err := c.ccloudDelete(cmd, deleteRoleBinding); err != nil {
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
			return errors.NewErrorWithSuggestions(
				fmt.Sprintf(httpStatusCodeErrorMsg, httpResp.StatusCode),
				httpStatusCodeSuggestions,
			)
		}

		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudDelete(cmd *cobra.Command, deleteRoleBinding *mdsv2.IamV2RoleBinding) error {
	roleBindings, err := c.V2Client.ListIamRoleBindings(deleteRoleBinding.GetCrnPattern(), deleteRoleBinding.GetPrincipal(), deleteRoleBinding.GetRoleName())
	if err != nil {
		return err
	}

	idx := slices.IndexFunc(roleBindings, func(roleBinding mdsv2.IamV2RoleBinding) bool {
		return roleBinding.GetCrnPattern() == deleteRoleBinding.GetCrnPattern()
	})
	if idx == -1 {
		return errors.NewErrorWithSuggestions(
			"failed to look up matching role binding",
			"To list role bindings, use `confluent iam rbac role-binding list`.",
		)
	}

	if err := deletion.ConfirmPromptYesOrNo(cmd, rbacPromptMsg); err != nil {
		return err
	}

	id := roleBindings[idx].GetId()
	deleteRoleBinding.SetId(id)

	_, err = c.V2Client.DeleteIamRoleBinding(id)
	return err
}

func (c *roleBindingCommand) confluentDelete(cmd *cobra.Command, options *roleBindingOptions) (*http.Response, error) {
	if err := deletion.ConfirmPromptYesOrNo(cmd, rbacPromptMsg); err != nil {
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
