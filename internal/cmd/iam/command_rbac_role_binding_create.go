package iam

import (
	"fmt"
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
		RunE:  c.create,
	}

	if c.cfg.IsCloudLogin() {
		cmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: `Grant the role "CloudClusterAdmin" to the principal "User:u-123456" in the environment "env-12345" for the cloud cluster "lkc-123456":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role CloudClusterAdmin --environment env-12345 --cloud-cluster lkc-123456",
			},
			examples.Example{
				Text: `Grant the role "ResourceOwner" to the principal "User:u-123456", in the environment "env-12345" for the Kafka cluster "lkc-123456" on the resource "Topic:my-topic":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --resource Topic:my-topic --environment env-12345 --kafka-cluster-id lkc-123456",
			},
			examples.Example{
				Text: `Grant the role "MetricsViewer" to service account "sa-123456":`,
				Code: "confluent iam rbac role-binding create --principal User:sa-123456 --role MetricsViewer",
			},
		)
	} else {
		cmd.Example = examples.BuildExampleString(
			examples.Example{
				Text: `Create a role binding for the principal permitting it produce to topic "my-topic":`,
				Code: "confluent iam rbac role-binding create --principal User:appSA --role DeveloperWrite --resource Topic:my-topic --kafka-cluster-id $KAFKA_CLUSTER_ID",
			},
		)
	}

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

	createRoleBinding, err := c.parseV2RoleBinding(cmd)
	if err != nil {
		return err
	}

	isCloud := c.cfg.IsCloudLogin()

	var httpResp *http.Response
	if isCloud {
		if isSchemaRegistryOrKsqlRoleBinding(createRoleBinding) {
			httpResp, err = c.ccloudCreate(options)
		} else {
			_, httpResp, err = c.V2Client.CreateIamRoleBinding(createRoleBinding)
		}
	} else {
		httpResp, err = c.confluentCreate(options)
	}
	if err != nil {
		return errors.CatchRequestNotValidMessageError(err, httpResp)
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
