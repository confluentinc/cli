package iam

import (
	"fmt"
	"io"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func (c *roleBindingCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete an existing role binding.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.delete),
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

	deleteRoleBinding, err := c.parseRoleBinding(cmd)
	if err != nil {
		return err
	}

	isCloud := c.cfg.IsCloudLogin()

	var httpResp *http.Response
	if isCloud {
		httpResp, err = c.ccloudDeleteV2(deleteRoleBinding)
		b, _ := io.ReadAll(httpResp.Body)
		fmt.Println(string(b))
		// resp, err = c.ccloudDelete(options)
	} else {
		httpResp, err = c.confluentDelete(options)
	}
	if err != nil {
		return err
	}

	// might be able to add catchers here, to print out more useful msgs. like 403: Either unauthorized to access role binding or role binding does not exist
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
	// do we distinguish resource nil or not nil? // probably not. It's all in the crn
	// maybe should not add the *? This is not listing everything but just one specific entry
	fmt.Println(*deleteRoleBinding.CrnPattern)
	resp, httpResp, err := c.V2Client.ListIamRoleBindingsNaive(deleteRoleBinding)
	if err != nil {
		return httpResp, err
	}
	if len(resp.Data) == 0 {
		return httpResp, errors.New("No matching role-bindings found.")
		// probably won't need this, the err code of resp will be caught anyway
	}
	id := *resp.Data[0].Id
	// are you supposed to delete more than one at a time?
	_, httpResp, err = c.V2Client.DeleteIamRoleBinding(id)
	return httpResp, err
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
