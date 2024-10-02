package flink

import (
	"fmt"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CERT_AUTHORITY_PATH" may be set in place of this flag.`)

	cmd.MarkFlagRequired("environment")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

// Run through all the pages until we get an empty page, in that case, return.
func getAllApplications(cmfClient *cmfsdk.APIClient, cmd *cobra.Command, environment string) ([]cmfsdk.Application, error) {
	applications := make([]cmfsdk.Application, 0)
	currentPageNumber := 0
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100

	pagingOptions := &cmfsdk.GetApplicationsOpts{
		Page: optional.NewInt32(int32(currentPageNumber)),
		Size: optional.NewInt32(pageSize),
	}

	for !done {
		applicationsPage, httpResponse, err := cmfClient.DefaultApi.GetApplications(cmd.Context(), environment, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list applications in the environment "%s": %s`, environment, parsedErr)
		}
		applications = append(applications, applicationsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(applicationsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return applications, nil
}

func (c *command) applicationList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applications, err := getAllApplications(cmfClient, cmd, environment)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			jobStatus, ok := app.Status["jobStatus"].(map[string]any)
			if !ok {
				jobStatus = map[string]any{}
			}
			envInApp, ok := app.Spec["environment"].(string)
			if !ok {
				envInApp = environment
			}
			list.Add(&flinkApplicationSummary{
				Name:        app.Metadata["name"].(string),
				Environment: envInApp,
				JobId:       jobStatus["jobId"].(string),
				JobState:    jobStatus["state"].(string),
			})
		}
		return list.Print()
	}
	// if the output format is not human, we serialize the output as it is (JSON or YAML)
	return output.SerializedOutput(cmd, applications)
}
