package iam

import (
	"fmt"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

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
				Code: "confluent iam rbac role-binding delete --principal User:u-123456 --role ResourceOwner --environment env-12345 --kafka-cluster-id lkc-123456 --resource Topic:my-topic",
			},
		)
	}

	cmd.Flags().String("role", "", "Role name of the existing role binding.")
	cmd.Flags().String("principal", "", "Qualified principal name associated with the role binding.")
	addClusterFlags(cmd, c.cfg.IsCloudLogin(), c.CLICommand)
	cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("principal")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func (c *roleBindingCommand) delete(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	isCloud := c.cfg.IsCloudLogin()

	var httpResp *http.Response
	if isCloud {
		if isSchemaRegistryOrKsqlRoleBinding(deleteRoleBinding) {
			httpResp, err = c.ccloudDelete(options)
		} else {
			deleteRoleBinding, err := c.parseV2RoleBinding(cmd)
			if err != nil {
				return err
			}
			httpResp, err = c.ccloudDeleteV2(deleteRoleBinding)
		}
	} else {
		httpResp, err = c.confluentDelete(options)
	}
	if err != nil {
		return errors.CatchRequestNotValidMessageError(err, httpResp)
	}

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, httpResp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	if isCloud {
		return c.displayCCloudCreateAndDeleteOutput(cmd, options)
	} else {
		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudDeleteV2(deleteRoleBinding *mdsv2.IamV2RoleBinding) (*http.Response, error) {
	resp, httpResp, err := c.V2Client.ListIamRoleBindings(deleteRoleBinding)
	if err != nil {
		return httpResp, err
	}
	roleBindingList := resp.Data

	for _, rolebinding := range roleBindingList {
		if *rolebinding.CrnPattern == *deleteRoleBinding.CrnPattern {
			_, httpResp, err = c.V2Client.DeleteIamRoleBinding(*rolebinding.Id)
			return httpResp, err
		}
	}
	return httpResp, errors.NewErrorWithSuggestions(errors.RoleBindingNotFoundFoundErrorMsg, errors.RoleBindingNotFoundFoundSuggestions)
}

func (c *roleBindingCommand) ccloudDelete(options *roleBindingOptions) (*http.Response, error) {
	if options.resource != "" {
		return c.MDSv2Client.RBACRoleBindingCRUDApi.RemoveRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequestV2)
	} else {
		return c.MDSv2Client.RBACRoleBindingCRUDApi.DeleteRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.scopeV2)
	}
}

func (c *roleBindingCommand) confluentDelete(options *roleBindingOptions) (*http.Response, error) {
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
