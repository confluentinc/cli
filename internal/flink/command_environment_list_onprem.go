package flink

import (
	"fmt"
	"io"
	"net/http"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
)

func (c *command) newEnvironmentListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Environments.",
		Args:  cobra.NoArgs,
		RunE:  c.environmentList,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) environmentList(cmd *cobra.Command, _ []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	environments, err := getAllEnvironments(cmfClient, cmd)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, env := range environments {
			list.Add(&flinkEnvironmentOut{
				Name:        env.Name,
				CreatedTime: env.CreatedTime.String(),
				UpdatedTime: env.UpdatedTime.String(),
			})
		}
		return list.Print()
	}
	return output.SerializedOutput(cmd, environments)
}

// Run through all the pages until we get an empty page, in that case, return.
func getAllEnvironments(cmfClient *cmfsdk.APIClient, cmd *cobra.Command) ([]cmfsdk.Environment, error) {
	environments := make([]cmfsdk.Environment, 0)
	page := 0
	lastPageEmpty := false

	pagingOptions := &cmfsdk.GetEnvironmentsOpts{
		Page: optional.NewInt32(int32(page)),
		// 100 is an arbitrary page size we've chosen.
		Size: optional.NewInt32(100),
	}

	for !lastPageEmpty {
		environmentsPage, httpResponse, err := cmfClient.DefaultApi.GetEnvironments(cmd.Context(), pagingOptions)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
				if httpResponse.Body != nil {
					defer httpResponse.Body.Close()
					respBody, parseError := io.ReadAll(httpResponse.Body)
					if parseError == nil {
						return nil, fmt.Errorf("failed to list environments: %s", string(respBody))
					}
				}
			}
			return nil, err
		}

		if environmentsPage.Items == nil || len(environmentsPage.Items) == 0 {
			lastPageEmpty = true
			break
		}
		environments = append(environments, environmentsPage.Items...)

		page += 1
		pagingOptions.Page = optional.NewInt32(int32(page))
	}

	return environments, nil

}
