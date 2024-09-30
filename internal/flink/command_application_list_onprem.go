package flink

import (
	"fmt"
	"io"
	"net/http"

	"github.com/antihax/optional"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	perrors "github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"
	"github.com/spf13/cobra"
)

func (c *unauthenticatedCommand) newApplicationListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink Applications.",
		Args:  cobra.NoArgs,
		RunE:  c.applicationList,
	}

	cmd.Flags().String("environment", "", "Name of the Environment to get the FlinkApplication from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

// Run through all the pages until we get an empty page, in that case, return.
func getAllApplications(cmfClient *cmfsdk.APIClient, cmd *cobra.Command, environment string) ([]cmfsdk.Application, error) {
	applications := make([]cmfsdk.Application, 0)
	page := 0
	lastPageEmpty := false

	pagingOptions := &cmfsdk.GetApplicationsOpts{
		Page: optional.NewInt32(int32(page)),
		// 100 is an arbitrary page size we've chosen.
		Size: optional.NewInt32(100),
	}

	for !lastPageEmpty {
		applicationsPage, httpResponse, err := cmfClient.DefaultApi.GetApplications(cmd.Context(), environment, pagingOptions)
		if err != nil {
			if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
				if httpResponse.Body != nil {
					defer httpResponse.Body.Close()
					respBody, parseError := io.ReadAll(httpResponse.Body)
					if parseError == nil {
						return nil, fmt.Errorf("failed to list applications in the environment \"%s\": %s", environment, string(respBody))
					}
				}
			}
			return nil, err
		}

		if applicationsPage.Items == nil || len(applicationsPage.Items) == 0 {
			lastPageEmpty = true
			break
		}
		applications = append(applications, applicationsPage.Items...)

		page += 1
		pagingOptions.Page = optional.NewInt32(int32(page))
	}

	return applications, nil

}

func (c *unauthenticatedCommand) applicationList(cmd *cobra.Command, _ []string) error {
	environment := getEnvironment(cmd)
	if environment == "" {
		return perrors.NewErrorWithSuggestions("environment name is required.", "You can use the --environment flag or set the default environment using `confluent flink environment use <name>` command.")
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
			jobStatus, ok := app.Status["jobStatus"].(map[string]interface{})
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
