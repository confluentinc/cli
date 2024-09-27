package flink

import (
	"errors"
	"io"

	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/spf13/cobra"
)

type deleteEnvironmentFailure struct {
	Environment string `human:"Environment" serialized:"environment"`
	Reason      string `human:"Reason" serialized:"reason"`
	StausCode   int    `human:"Status Code" serialized:"status_code"`
}

func (c *unauthenticatedCommand) newEnvironmentDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name-1> [name-2] ... [name-n]",
		Short: "Delete one or more Flink Environments.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.environmentDelete,
	}

	return cmd
}

func (c *unauthenticatedCommand) environmentDelete(cmd *cobra.Command, args []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	deleteFunc := func(name string) error {
		httpResp, err := cmfClient.DefaultApi.DeleteEnvironment(cmd.Context(), name)
		if err != nil && httpResp != nil {
			if httpResp.Body != nil {
				defer httpResp.Body.Close()
				respBody, parseError := io.ReadAll(httpResp.Body)
				if parseError == nil {
					return errors.New(string(respBody))
				}
			}
		}
		return err
	}

	_, err = deletion.Delete(args, deleteFunc, resource.OnPremFlinkEnvrionment)
	return err
}
