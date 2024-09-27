package flink

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"

	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
)

func (c *unauthenticatedCommand) newApplicationWebUiCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forward-web-ui <name>",
		Short: "Forward the web UI of a Flink Application.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.applicationForwardWebUi,
	}
	cmd.Flags().String("environment", "", "Name of the environment for the Flink Application.")
	cmd.Flags().Int("port", 0, "Port to forward the web UI to. If not provided, a random, OS-assigned port will be used.")
	return cmd
}

func (c *unauthenticatedCommand) applicationForwardWebUi(cmd *cobra.Command, args []string) error {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("url is required")
	}

	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environment == "" {
		return fmt.Errorf("environment name is required")
	}

	port, err := cmd.Flags().GetInt("port")
	if err != nil {
		return err
	}
	if port < 0 {
		return fmt.Errorf("port must be a positive integer")
	}

	cmfClient, err := c.GetCmfClient()
	if err != nil {
		return err
	}

	// Get the name of the application
	applicationName := args[0]
	_, httpResponse, err := cmfClient.DefaultApi.GetApplication(cmd.Context(), environment, applicationName, nil)

	// check if the application exists
	if httpResponse != nil && httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("application \"%s\" does not exist in the environment \"%s\" or environment \"%s\" does not exist", applicationName, environment, environment)
	}

	client := &http.Client{}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handleRequest(w, r, url, environment, applicationName, client)
	})

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to create the local listener: %s", err)
	}

	// We log just before starting because Serve() blocks, so we can't print after it.
	output.Printf(false, "Starting web UI at http://localhost:%+v/ ... (Press ^C to stop)", listener.Addr().(*net.TCPAddr).Port)
	http.Serve(listener, nil)
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
		http.Error(userResponseWriter, fmt.Sprintf("failed to forward the web-ui: %s", err), http.StatusInternalServerError)
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
		http.Error(userResponseWriter, fmt.Sprintf("failed to return response from the web-ui: %s", err), http.StatusInternalServerError)
		return
	}
	userResponseWriter.Write(resBody)
}
