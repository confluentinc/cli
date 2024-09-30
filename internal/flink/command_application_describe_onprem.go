package flink

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/spf13/cobra"

	perrors "github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *unauthenticatedCommand) newApplicationDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <name>",
		Short: "Describe a Flink Application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationDescribe,
	}

	cmd.Flags().String("environment", "", "Name of the Environment to describe the FlinkApplication in.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", "Path to client private key, include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_KEY_PATH.")
	cmd.Flags().String("client-cert-path", "", "Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Flag can also be set via CONFLUENT_CMF_CLIENT_CERT_PATH.")
	cmd.Flags().String("certificate-authority-path", "", "Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Flag can also be set via CONFLUENT_CERT_AUTHORITY_PATH.")

	return cmd
}

func (c *unauthenticatedCommand) applicationDescribe(cmd *cobra.Command, args []string) error {
	environment := getEnvironment(cmd)
	if environment == "" {
		return perrors.NewErrorWithSuggestions("environment name is required", "You can use the --environment flag or set the default environment using `confluent flink environment use <name>` command")
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the application to be retrieved
	applicationName := args[0]
	cmfApplication, httpResponse, err := cmfClient.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)

	if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
		// Read response body if any
		respBody := []byte{}
		var parseError error
		if httpResponse.Body != nil {
			defer httpResponse.Body.Close()
			respBody, parseError = ioutil.ReadAll(httpResponse.Body)
			if parseError != nil {
				respBody = []byte(fmt.Sprintf("failed to read response body: %s", parseError))
			}
		}
		// Start checking the possible status codes
		switch httpResponse.StatusCode {
		case http.StatusNotFound:
			return fmt.Errorf("application \"%s\" not found %s", applicationName, string(respBody))
		case http.StatusInternalServerError:
			return fmt.Errorf("internal server error while describing application \"%s\": %s", applicationName, string(respBody))
		default:
			return fmt.Errorf("failed to describe application \"%s\": %s", applicationName, err)
		}
	}

	if err != nil {
		return fmt.Errorf("failed to describe application \"%s\" in the environment \"%s\": %s", applicationName, environment, err)
	}

	table := output.NewTable(cmd)
	var metadataBytes, specBytes, statusBytes []byte
	metadataBytes, err = json.Marshal(cmfApplication.Metadata)
	specBytes, err = json.Marshal(cmfApplication.Spec)
	statusBytes, err = json.Marshal(cmfApplication.Status)

	table.Add(&flinkApplicationOutput{
		ApiVersion: cmfApplication.ApiVersion,
		Kind:       cmfApplication.Kind,
		Metadata:   string(metadataBytes),
		Spec:       string(specBytes),
		Status:     string(statusBytes),
	})
	return table.Print()
}
