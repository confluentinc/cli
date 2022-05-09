package iam

import (
	"fmt"
	"io"
	"net/http"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/spf13/cobra"
)

func (c *roleBindingCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role binding.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.create),
	}

	example := examples.Example{
		Text: `Create a role binding for the principal permitting it produce to the "users" topic.`,
		Code: "confluent iam rbac role-binding create --principal User:appSA --role DeveloperWrite --resource Topic:users --kafka-cluster-id $KAFKA_CLUSTER_ID",
	}
	if c.cfg.IsCloudLogin() {
		example.Code = "confluent iam rbac role-binding create --principal User:u-ab1234 --role DeveloperWrite --resource Topic:users --cloud-cluster lkc-ab123 --environment env-abcde"
	}
	cmd.Example = examples.BuildExampleString(example)

	cmd.Flags().String("role", "", "Role name of the new role binding.")
	cmd.Flags().String("principal", "", "Qualified principal name for the role binding.")

	addClusterFlags(cmd, c.cfg.IsCloudLogin(), c.CLICommand)

	cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")

	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("role")
	_ = cmd.MarkFlagRequired("principal")

	return cmd
}

func (c *roleBindingCommand) create(cmd *cobra.Command, _ []string) error {
	options, err := c.parseCommon(cmd)
	if err != nil {
		return err
	}

	createRoleBinding, err := c.parseRoleBinding(cmd)
	if err != nil {
		return err
	}
	fmt.Println(*createRoleBinding.Principal)
	fmt.Println(*createRoleBinding.RoleName)
	fmt.Println(*createRoleBinding.CrnPattern)

	isCloud := c.cfg.IsCloudLogin()

	var httpResp *http.Response
	if isCloud {
		// resp, err = c.ccloudCreate(options)
		_, httpResp, err = c.V2Client.CreateIamRoleBinding(createRoleBinding)
		b, _ := io.ReadAll(httpResp.Body)
		fmt.Println(string(b))
	} else {
		httpResp, err = c.confluentCreate(options)
	}
	if err != nil {
		return err
	}

	if httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, httpResp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	if isCloud {
		return c.displayCCloudCreateAndDeleteOutput(cmd, options)
	} else {
		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudCreate(options *roleBindingOptions) (*http.Response, error) {
	if options.resource != "" {
		return c.MDSv2Client.RBACRoleBindingCRUDApi.AddRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequestV2)
	} else {
		return c.MDSv2Client.RBACRoleBindingCRUDApi.AddRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.scopeV2)
	}
}

func (c *roleBindingCommand) confluentCreate(options *roleBindingOptions) (*http.Response, error) {
	if options.resource != "" {
		return c.MDSClient.RBACRoleBindingCRUDApi.AddRoleResourcesForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.resourcesRequest)
	} else {
		return c.MDSClient.RBACRoleBindingCRUDApi.AddRoleForPrincipal(
			c.createContext(),
			options.principal,
			options.role,
			options.mdsScope)
	}
}
