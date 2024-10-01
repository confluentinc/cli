package flink

import (
	"fmt"
	"io"
	"net/http"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	currentPageNumber := 0
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100

	pagingOptions := &cmfsdk.GetEnvironmentsOpts{
		Page: optional.NewInt32(int32(currentPageNumber)),
		// 100 is an arbitrary page size we've chosen.
		Size: optional.NewInt32(pageSize),
	}

	for !done {
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

		environments = append(environments, environmentsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(environmentsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return environments, nil
}
