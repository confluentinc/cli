package flink

import (
	"fmt"
	"io"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
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

	cmd.Flags().StringP("environment", "e", "", "Name of the Environment to get the FlinkApplication from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")

	cmd.MarkFlagRequired("environment")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *unauthenticatedCommand) applicationList(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environment == "" {
		return errors.NewErrorWithSuggestions("environment is required", "set the environment with --environment flag")
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	applicationsPage, httpResponse, err := cmfClient.DefaultApi.GetApplications(cmd.Context(), environment, nil)
	if err != nil {
		if httpResponse != nil && httpResponse.StatusCode != 200 {
			if httpResponse.Body != nil {
				defer httpResponse.Body.Close()
				respBody, parseError := io.ReadAll(httpResponse.Body)
				if parseError == nil {
					return fmt.Errorf("failed to list applications in the environment \"%s\": %s", environment, string(respBody))
				}
			}
		}
		return err
	}

	var list []cmfsdk.Application
	applications := append(list, applicationsPage.Items...)

	if output.GetFormat(cmd) == output.Human {
		list := output.NewList(cmd)
		for _, app := range applications {
			jobStatus := app.Status["jobStatus"].(map[string]interface{})
			envInApp, ok := app.Spec["environment"].(string)
			if !ok {
				envInApp = environment
			}
			list.Add(&flinkApplicationOut{
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
