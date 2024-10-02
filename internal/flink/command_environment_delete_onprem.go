package flink

import (
	"net/http"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newEnvironmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink environments.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.environmentDelete,
	}

	pcmd.AddForceFlag(cmd)
	return cmd
}

func (c *command) environmentDelete(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	existenceFunc := func(name string) bool {
		_, httpResp, err := cmfClient.DefaultApi.GetEnvironment(cmd.Context(), name)
		return err == nil && httpResp.StatusCode == http.StatusOK
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.FlinkEnvironment); err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		httpResp, err := cmfClient.DefaultApi.DeleteEnvironment(cmd.Context(), name)
		return parseSdkError(httpResp, err)
	}

	_, err = deletion.Delete(args, deleteFunc, resource.FlinkEnvironment)
	return err
}
