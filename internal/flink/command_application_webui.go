package flink

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/flink"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newApplicationWebUiForwardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "web-ui-forward <name>",
		Short: "Forward the web UI of a Flink application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationWebUiForward,
	}

	cmd.Flags().String("environment", "", "Name of the environment to delete the Flink application from.")
	cmd.Flags().String("url", "", `Base URL of the Confluent Manager for Apache Flink (CMF). Environment variable "CONFLUENT_CMF_URL" may be set in place of this flag.`)
	cmd.Flags().String("client-key-path", "", `Path to client private key for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_KEY_PATH" may be set in place of this flag.`)
	cmd.Flags().String("client-cert-path", "", `Path to client cert to be verified by Confluent Manager for Apache Flink. Include for mTLS authentication. Environment variable "CONFLUENT_CMF_CLIENT_CERT_PATH" may be set in place of this flag.`)
	cmd.Flags().String("certificate-authority-path", "", `Path to a PEM-encoded Certificate Authority to verify the Confluent Manager for Apache Flink connection. Environment variable "CONFLUENT_CMF_CERTIFICATE_AUTHORITY_PATH" may be set in place of this flag.`)
	cmd.Flags().Uint16("port", 0, "Port to forward the web UI to. If not provided, a random, OS-assigned port will be used.")

	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("url"))

	return cmd
}

func (c *command) applicationWebUiForward(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	port, err := cmd.Flags().GetUint16("port")
	if err != nil {
		return err
	}

	cmfClient, err := c.GetCmfClient(cmd)
	if err != nil {
		return err
	}

	// Get the name of the application and check for its existence
	applicationName := args[0]
	_, err = cmfClient.DescribeApplication(cmd.Context(), environment, applicationName)
	if err != nil {
		return fmt.Errorf(`application "%s" does not exist in the environment "%s" or environment "%s" does not exist`, applicationName, environment, environment)
	}

	restFlags, err := flink.ResolveOnPremCmfRestFlags(cmd)
	if err != nil {
		return err
	}

	client, err := flink.NewCmfRestHttpClient(restFlags)
	if err != nil {
		return err
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, url, environment, applicationName, client)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to create the local listener: %s", err)
	}

	// We log just before starting because Serve() blocks, so we can't print after it.
	output.Printf(false, "Starting web UI at http://localhost:%+v/ ... (Press Ctrl-C to stop)\n", listener.Addr().(*net.TCPAddr).Port)

	_ = http.Serve(listener, nil)
	return nil
}

func handleRequest(userResponseWriter http.ResponseWriter, userRequest *http.Request, url, environmentName, applicationName string, client *http.Client) {
	body, err := io.ReadAll(userRequest.Body)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("Failed to read request body: %s", err), http.StatusInternalServerError)
		return
	}

	newUrl := fmt.Sprintf("%s/cmf/api/v1/environments/%s/applications/%s/flink-web-ui%s", url, environmentName, applicationName, userRequest.RequestURI)
	reqToCmf, err := http.NewRequest(userRequest.Method, newUrl, bytes.NewReader(body))
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to forward the web UI: %s", err), http.StatusInternalServerError)
		return
	}
	reqToCmf.Header = userRequest.Header

	resFromCmf, err := client.Do(reqToCmf)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to forward the request: %s", err), http.StatusInternalServerError)
		return
	}
	defer resFromCmf.Body.Close()

	// Copy response headers - this includes content type.
	for key, values := range resFromCmf.Header {
		for _, value := range values {
			userResponseWriter.Header().Set(key, value)
		}
	}
	userResponseWriter.WriteHeader(resFromCmf.StatusCode)

	// Copy response body.
	resBody, err := io.ReadAll(resFromCmf.Body)
	if err != nil {
		http.Error(userResponseWriter, fmt.Sprintf("failed to return response from the web UI: %s", err), http.StatusInternalServerError)
		return
	}
	_, err = userResponseWriter.Write(resBody)
	if err != nil {
		output.ErrPrintf(false, "Failed to write response body: %s", err)
	}
}
