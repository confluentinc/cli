package flink

import (
	"fmt"
	"io"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

func (c *command) newEnvironmentListOnPremCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Environments.",
		RunE:  c.listEnvironmentsOnPrem,
	}

	pcmd.AddOutputFlag(cmd)
	return cmd
}

func (c *command) listEnvironmentsOnPrem(cmd *cobra.Command, _ []string) error {
	cmfREST, err := c.GetCmfREST()
	if err != nil {
		return err
	}

	environmentsPage, httpResponse, err := cmfREST.Client.DefaultApi.GetEnvironments(cmd.Context(), nil)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			if httpResponse.Body != nil {
				defer httpResponse.Body.Close()
				respBody, parseError := io.ReadAll(httpResponse.Body)
				if parseError == nil {
					return fmt.Errorf("failed to list environments: %s", string(respBody))
				}
			}
		}
		return err
	}

	var list []cmfsdk.GetEnvironment
	environments := append(list, environmentsPage.Items...)

	// TODO: Add pagination support once the API supports it
	if len(environments) == 0 {
		return fmt.Errorf("no environments found")
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, env := range environments {
			list.Add(&flinkEnvironmentOut{
				Name:            env.Name,
				DefaultStrategy: env.DefaultStrategy,
				CreatedTime:     env.CreatedTime.String(),
				UpdatedTime:     env.UpdatedTime.String(),
			})
		}
		return list.Print()
	}
	return output.SerializedOutput(cmd, environments)
}
