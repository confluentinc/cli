package iam

import (
	"fmt"
	"net/http"

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

	if c.cfg.IsCloudLogin() {
		cmd.Flags().String("cloud-cluster", "", "Cloud cluster ID for the role binding.")
		cmd.Flags().String("environment", "", "Environment ID for scope of role-binding delete.")
		cmd.Flags().Bool("current-env", false, "Use current environment ID for scope.")
		if c.ccloudRbacDataplaneEnabled {
			cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
			cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
			cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		}
	} else {
		cmd.Flags().Bool("prefix", false, "Whether the provided resource name is treated as a prefix pattern.")
		cmd.Flags().String("resource", "", "Qualified resource name for the role binding.")
		cmd.Flags().String("kafka-cluster-id", "", "Kafka cluster ID for the role binding.")
		cmd.Flags().String("schema-registry-cluster-id", "", "Schema Registry cluster ID for the role binding.")
		cmd.Flags().String("ksql-cluster-id", "", "ksqlDB cluster ID for the role binding.")
		cmd.Flags().String("connect-cluster-id", "", "Kafka Connect cluster ID for the role binding.")
		cmd.Flags().String("cluster-name", "", "Cluster name to uniquely identify the cluster for role binding listings.")
		pcmd.AddContextFlag(cmd, c.CLICommand)
	}

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

	var resp *http.Response
	if isCloud {
		resp, err = c.ccloudDelete(options)
	} else {
		resp, err = c.confluentDelete(options)
	}
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.HTTPStatusCodeErrorMsg, resp.StatusCode), errors.HTTPStatusCodeSuggestions)
	}

	if isCloud {
		return c.displayCCloudCreateAndDeleteOutput(cmd, options)
	} else {
		return displayCreateAndDeleteOutput(cmd, options)
	}
}

func (c *roleBindingCommand) ccloudDelete(options *roleBindingOptions) (*http.Response, error) {
	if c.ccloudRbacDataplaneEnabled && options.resource != "" {
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
