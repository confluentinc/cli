package iam

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *roleBindingCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a role binding.",
		Args:  cobra.NoArgs,
		RunE:  c.create,
	}

	var exs []examples.Example

	if c.cfg.IsCloudLogin() {
		exs = append(exs,
			examples.Example{
				Text: `Grant the role "CloudClusterAdmin" to the principal "User:u-123456" in the environment "env-12345" for the cloud cluster "lkc-123456":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role CloudClusterAdmin --environment env-12345 --cloud-cluster lkc-123456",
			},
			examples.Example{
				Text: `Grant the role "ResourceOwner" to the principal "User:u-123456", in the environment "env-12345" for the Kafka cluster "lkc-123456" on the resource "Topic:my-topic":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --resource Topic:my-topic --environment env-12345 --kafka-cluster lkc-123456",
			},
			examples.Example{
				Text: `Grant the role "MetricsViewer" to service account "sa-123456":`,
				Code: "confluent iam rbac role-binding create --principal User:sa-123456 --role MetricsViewer",
			},
			examples.Example{
				Text: `Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects for Schema Registry cluster "lsrc-123456" in environment "env-12345":`,
				Code: `confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject:*"`,
			},
			examples.Example{
				Text: `Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" for the Schema Registry cluster "lsrc-123456" in the environment "env-12345":`,
				Code: `confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject:test"`,
			},
			examples.Example{
				Text: `Grant the "ResourceOwner" role to principal "User:u-123456" and all subjects in schema context "schema_context" for Schema Registry cluster "lsrc-123456" in the environment "env-12345":`,
				Code: `confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:*"`,
			},
			examples.Example{
				Text: `Grant the "ResourceOwner" role to principal "User:u-123456" and subject "test" in schema context "schema_context" for Schema Registry "lsrc-123456" in the environment "env-12345":`,
				Code: `confluent iam rbac role-binding create --principal User:u-123456 --role ResourceOwner --environment env-12345 --schema-registry-cluster lsrc-123456 --resource "Subject::.schema_context:test"`,
			},
			examples.Example{
				Text: `Grant the role "FlinkDeveloper" to the principal "User:u-123456", in the environment "env-12345":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role FlinkDeveloper --environment env-12345",
			},
			examples.Example{
				Text: `Grant the role "FlinkDeveloper" to the principal "User:u-123456", in the environment "env-12345" for the compute pool "lfcp-123456" in the Flink region "us-east-2":`,
				Code: "confluent iam rbac role-binding create --principal User:u-123456 --role FlinkDeveloper --environment env-12345 --flink-region aws.us-east-2 --resource ComputePool:lfcp-123456",
			},
		)
	} else {
		exs = append(exs,
			examples.Example{
				Text: `Create a role binding for the principal permitting it produce to topic "my-topic":`,
				Code: "confluent iam rbac role-binding create --principal User:appSA --role DeveloperWrite --resource Topic:my-topic --kafka-cluster $KAFKA_CLUSTER_ID",
			},
		)
	}

	cmd.Example = examples.BuildExampleString(exs...)

	cmd.Flags().String("role", "", "Role name of the new role binding.")
	cmd.Flags().String("principal", "", "Qualified principal name for the role binding.")
	addClusterFlags(cmd, c.cfg, c.CLICommand)
	cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
	cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("role"))
	cobra.CheckErr(cmd.MarkFlagRequired("principal"))

	return cmd
}

func (c *roleBindingCommand) create(cmd *cobra.Command, _ []string) error {
	isCloud := c.cfg.IsCloudLogin()

	if isCloud {
		createRoleBinding, err := c.parseV2RoleBinding(cmd)
		if err != nil {
			return err
		}

		resp, err := c.V2Client.CreateIamRoleBinding(createRoleBinding)
		if err != nil {
			return err
		}
		createRoleBinding.SetId(resp.GetId())

		return c.displayCCloudCreateAndDeleteOutput(cmd, createRoleBinding)
	} else {
		options, err := c.parseCommon(cmd)
		if err != nil {
			return err
		}
		httpResp, err := c.confluentCreate(options)
		if err != nil {
			return err
		}

		if httpResp != nil && httpResp.StatusCode != http.StatusOK && httpResp.StatusCode != http.StatusCreated && httpResp.StatusCode != http.StatusNoContent {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, httpResp.StatusCode), errors.HTTPStatusCodeSuggestions)
		}

		return displayCreateAndDeleteOutput(cmd, options)
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
