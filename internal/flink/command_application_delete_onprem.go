package flink

import (
	"fmt"
	"io"
	"strings"

	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

type deleteApplicationFailure struct {
	Application string `human:"Application" serialized:"application"`
	Environment string `human:"Environment" serialized:"environment"`
	Reason      string `human:"Reason" serialized:"reason"`
	StausCode   int    `human:"Status Code" serialized:"status_code"`
}

func (c *command) newApplicationDeleteCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <name>[, <name>*]",
		Short: "Delete given Flink Application(s).",
		Long:  "Delete given Flink Application(s). In case you want to delete multiple applications, the names should be separated by a comma and the applications should belong to the same environment.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.deleteApplicationOnPrem,
	}

	return cmd
}

func (c *command) deleteApplicationOnPrem(cmd *cobra.Command, _ []string) error {
	cmfREST, err := c.GetCmfRest()
	if err != nil {
		return err
	}

	environmentName, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environmentName == "" {
		fmt.Errorf("Environment name is required")
		return nil
	}

	// Range over the arguments and split them by comma
	args := cmd.Flags().Args()
	applicationNames := make([]string, 0)
	applicationNames = append(applicationNames, strings.Split(args[0], ",")...)
	// create a list of failed application deletions
	failedDeletions := make([]deleteApplicationFailure, 0)
	successfulDeletions := make([]string, 0)
	for _, appName := range applicationNames {
		appName = strings.TrimSpace(appName) // Clean up whitespace if any
		if appName != "" {
			httpResponse, err := cmfREST.Client.DefaultApi.DeleteApplication(cmd.Context(), environmentName, appName)
			if err != nil {
				if httpResponse != nil && httpResponse.StatusCode != 200 {
					if httpResponse.Body != nil {
						defer httpResponse.Body.Close()
						respBody, parseError := io.ReadAll(httpResponse.Body)
						if parseError == nil {
							failedDeletions = append(failedDeletions, deleteApplicationFailure{
								Application: appName,
								Environment: environmentName,
								Reason:      string(respBody),
								StausCode:   httpResponse.StatusCode,
							})
						}
					}
				}
			} else {
				successfulDeletions = append(successfulDeletions, appName)
			}
		}
	}
	if len(successfulDeletions) > 0 {
		fmt.Printf("Application(s) deleted successfully: %s\n\n", strings.Join(successfulDeletions, ", "))
	}
	if len(failedDeletions) == 0 {
		return nil
	}

	fmt.Printf("failed to delete the following application(s):\n")
	failedDeletionsList := output.NewList(cmd)
	for _, failedDeletion := range failedDeletions {
		failedDeletionsList.Add(&failedDeletion)
	}
	return failedDeletionsList.Print()
}
