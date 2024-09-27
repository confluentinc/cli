package flink

import (
	"fmt"
	"io"
	"strings"

	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

type deleteEnvironmentFailure struct {
	Environment string `human:"Environment" serialized:"environment"`
	Reason      string `human:"Reason" serialized:"reason"`
	StausCode   int    `human:"Status Code" serialized:"status_code"`
}

func (c *unauthenticatedCommand) newEnvironmentDeleteommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>[, <name>*]",
		Short: "Delete given Flink Environment(s).",
		Long:  "Delete given Flink Environment(s). In case you want to delete multiple environments, the names should be separated by a comma.",

		Args: cobra.MinimumNArgs(1),
		RunE: c.environmentDelete,
	}

	return cmd
}

func (c *unauthenticatedCommand) environmentDelete(cmd *cobra.Command, _ []string) error {
	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Range over the arguments and split them by comma
	args := cmd.Flags().Args()
	environmentNames := make([]string, 0)
	environmentNames = append(environmentNames, strings.Split(args[0], ",")...)
	// create a list of failed environment deletions
	failedDeletions := make([]deleteEnvironmentFailure, 0)
	successfulDeletions := make([]string, 0)
	for _, envName := range environmentNames {
		envName = strings.TrimSpace(envName) // Clean up whitespace if any
		if envName != "" {
			httpResponse, err := cmfClient.DefaultApi.DeleteEnvironment(cmd.Context(), envName)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode != 200 {
					if httpResponse.Body != nil {
						defer httpResponse.Body.Close()
						respBody, parseError := io.ReadAll(httpResponse.Body)
						if parseError == nil {
							failedDeletions = append(failedDeletions, deleteEnvironmentFailure{
								Environment: envName,
								Reason:      string(respBody),
								StausCode:   httpResponse.StatusCode,
							})
						}
					}
				}
			} else {
				successfulDeletions = append(successfulDeletions, envName)
			}
		}
	}
	if len(successfulDeletions) > 0 {
		fmt.Printf("Environment(s) deleted successfully: %s\n\n", strings.Join(successfulDeletions, ", "))
	}
	if len(failedDeletions) == 0 {
		return nil
	}
	fmt.Printf("failed to delete the following environment(s):\n")
	failedDeletionsList := output.NewList(cmd)
	for _, failedDeletion := range failedDeletions {
		failedDeletionsList.Add(&failedDeletion)
	}
	return failedDeletionsList.Print()
}
