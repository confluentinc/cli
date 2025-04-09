package flink

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

type OnPremCMFRestFlagValues struct {
	url            string
	caCertPath     string
	clientCertPath string
	clientKeyPath  string
}

type CmfRestClient struct {
	*cmfsdk.APIClient
}

func NewCmfRestHttpClient(restFlags *OnPremCMFRestFlagValues) (*http.Client, error) {
	var err error
	httpClient := utils.DefaultClient()

	// If caCertPath is not provided via flag, check if it is set in the environment
	if restFlags.caCertPath == "" {
		restFlags.caCertPath = os.Getenv(auth.ConfluentPlatformCmfCertificateAuthorityPath)
	}
	// If we find a caCertPath, we will use it to create the client using the custom certificate authority
	if restFlags.caCertPath != "" {
		httpClient, err = utils.CustomCAAndClientCertClient(restFlags.caCertPath, restFlags.clientCertPath, restFlags.clientKeyPath)
		if err != nil {
			return nil, err
		}
	} else if restFlags.clientCertPath != "" && restFlags.clientKeyPath != "" {
		httpClient, err = utils.CustomCAAndClientCertClient("", restFlags.clientCertPath, restFlags.clientKeyPath)
		if err != nil {
			return nil, err
		}
	}

	return httpClient, nil
}

func NewCmfRestClient(cfg *cmfsdk.Configuration, restFlags *OnPremCMFRestFlagValues) (*CmfRestClient, error) {
	var err error
	cmfRestClient := &CmfRestClient{}

	// Set the base path if it's not set (it'll be already set in case of tests).
	if cfg.BasePath == "" {
		if restFlags.url == "" {
			return nil, errors.NewErrorWithSuggestions("url is required", "Specify a URL with `--url` or set the variable \"CONFLUENT_CMF_URL\" in place of this flag.")
		}
		cfg.BasePath, err = url.JoinPath(restFlags.url, "/cmf/api/v1")
		if err != nil {
			return nil, err
		}
	}

	cfg.HTTPClient, err = NewCmfRestHttpClient(restFlags)
	if err != nil {
		return nil, err
	}

	client := cmfsdk.NewAPIClient(cfg)
	cmfRestClient.APIClient = client
	return cmfRestClient, nil
}

func ResolveOnPremCmfRestFlags(cmd *cobra.Command) (*OnPremCMFRestFlagValues, error) {
	url, err := cmd.Flags().GetString("url")
	if err != nil {
		return nil, err
	}
	if url == "" {
		url = os.Getenv(auth.ConfluentPlatformCmfURL)
	}

	certificateAuthorityPath, err := cmd.Flags().GetString("certificate-authority-path")
	if err != nil {
		return nil, err
	}
	if certificateAuthorityPath == "" {
		certificateAuthorityPath = os.Getenv(auth.ConfluentPlatformCmfCertificateAuthorityPath)
	}

	clientCertPath, err := cmd.Flags().GetString("client-cert-path")
	if err != nil {
		return nil, err
	}
	if clientCertPath == "" {
		clientCertPath = os.Getenv(auth.ConfluentPlatformCmfClientCertPath)
	}

	clientKeyPath, err := cmd.Flags().GetString("client-key-path")
	if err != nil {
		return nil, err
	}
	if clientKeyPath == "" {
		clientKeyPath = os.Getenv(auth.ConfluentPlatformCmfClientKeyPath)
	}

	values := &OnPremCMFRestFlagValues{
		url:            url,
		caCertPath:     certificateAuthorityPath,
		clientCertPath: clientCertPath,
		clientKeyPath:  clientKeyPath,
	}
	return values, nil
}

// Create an application in the specified environment.
// Internally, since the call for Create and Update is the same, we check if the environment doesn't contain said application before creation.
func (cmfClient *CmfRestClient) CreateApplication(ctx context.Context, environment string, application cmfsdk.Application) (cmfsdk.Application, error) {
	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, _ := cmfClient.DefaultApi.GetApplication(ctx, environment, applicationName)
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusOK {
		return cmfsdk.Application{}, fmt.Errorf(`application "%s" already exists in the environment "%s"`, applicationName, environment)
	}

	outputApplication, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateApplication(ctx, environment, application)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Application{}, fmt.Errorf(`failed to create application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}
	return outputApplication, nil
}

func (cmfClient *CmfRestClient) DeleteApplication(ctx context.Context, environment, application string) error {
	httpResp, err := cmfClient.DefaultApi.DeleteApplication(ctx, environment, application)
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeApplication(ctx context.Context, environment, application string) (cmfsdk.Application, error) {
	cmfApplication, httpResponse, err := cmfClient.DefaultApi.GetApplication(ctx, environment, application)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Application{}, fmt.Errorf(`failed to describe application "%s" in the environment "%s": %s`, application, environment, parsedErr)
	}
	return cmfApplication, nil
}

func (cmfClient *CmfRestClient) ListApplications(ctx context.Context, environment string) ([]cmfsdk.Application, error) {
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
		applicationsPage, httpResponse, err := cmfClient.DefaultApi.GetApplications(ctx, environment, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list applications in the environment "%s": %s`, environment, parsedErr)
		}
		applications = append(applications, applicationsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(applicationsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return applications, nil
}

// Update an application in the specified environment.
// Internally, since the call for Create and Update is the same, we check if the environment contains said application before updation.
func (cmfClient *CmfRestClient) UpdateApplication(ctx context.Context, environment string, application cmfsdk.Application) (cmfsdk.Application, error) {
	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, err := cmfClient.DefaultApi.GetApplication(ctx, environment, applicationName)
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
		return cmfsdk.Application{}, fmt.Errorf(`application "%s" does not exist in the environment "%s"`, applicationName, environment)
	} else if httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		// Any failure other than 404 is an error in the response and shouldn't be treated as the application not existing.
		parsedErr := parseSdkError(httpResponse, err)
		return cmfsdk.Application{}, fmt.Errorf(`failed to update application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}

	outputApplication, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateApplication(ctx, environment, application)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Application{}, fmt.Errorf(`failed to update application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}
	return outputApplication, nil
}

// Create an environment.
// Internally, since the call for Create and Update is the same, we check if the environment exists before creation.
func (cmfClient *CmfRestClient) CreateEnvironment(ctx context.Context, postEnvironment cmfsdk.PostEnvironment) (cmfsdk.Environment, error) {
	environmentName := postEnvironment.Name
	_, httpResponse, _ := cmfClient.DefaultApi.GetEnvironment(ctx, environmentName)
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusOK {
		return cmfsdk.Environment{}, fmt.Errorf(`environment "%s" already exists`, environmentName)
	}

	outputEnvironment, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateEnvironment(ctx, postEnvironment)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to create environment "%s": %s`, environmentName, parsedErr)
	}
	return outputEnvironment, nil
}

func (cmfClient *CmfRestClient) DeleteEnvironment(ctx context.Context, environment string) error {
	httpResp, err := cmfClient.DefaultApi.DeleteEnvironment(ctx, environment)
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeEnvironment(ctx context.Context, environment string) (cmfsdk.Environment, error) {
	cmfEnvironment, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(ctx, environment)

	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to describe environment "%s": %s`, environment, parsedErr)
	}

	return cmfEnvironment, nil
}

// Run through all the pages until we get an empty page, in that case, return.
func (cmfClient *CmfRestClient) ListEnvironments(ctx context.Context) ([]cmfsdk.Environment, error) {
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
		environmentsPage, httpResponse, err := cmfClient.DefaultApi.GetEnvironments(ctx, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf("failed to list environments: %s", parsedErr)
		}

		environments = append(environments, environmentsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(environmentsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return environments, nil
}

// Create an environment.
// Internally, since the call for Create and Update is the same, we check if the environment exists before updation.
func (cmfClient *CmfRestClient) UpdateEnvironment(ctx context.Context, postEnvironment cmfsdk.PostEnvironment) (cmfsdk.Environment, error) {
	environmentName := postEnvironment.Name
	_, httpResponse, err := cmfClient.DefaultApi.GetEnvironment(ctx, environmentName)
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
		return cmfsdk.Environment{}, fmt.Errorf(`environment "%s" does not exist`, environmentName)
	} else if httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		// Any failure other than 404 is an error in the response and shouldn't be treated as the environment not existing.
		parsedErr := parseSdkError(httpResponse, err)
		return cmfsdk.Environment{}, fmt.Errorf(`failed to update environment "%s": %s`, environmentName, parsedErr)
	}

	outputEnvironment, httpResponse, err := cmfClient.DefaultApi.CreateOrUpdateEnvironment(ctx, postEnvironment)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to update environment "%s": %s`, environmentName, parsedErr)
	}
	return outputEnvironment, nil
}

// Returns the next page number and whether we need to fetch more pages or not.
func extractPageOptions(receivedItemsLength int, currentPageNumber int) (int, bool) {
	if receivedItemsLength == 0 {
		return currentPageNumber, true
	}
	return currentPageNumber + 1, false
}

// Creates a rich error message from the HTTP response and the SDK error if possible.
func parseSdkError(httpResp *http.Response, sdkErr error) error {
	// If there's an error, and the httpResp is populated, it may contain a more detailed error message.
	// If there's nothing in the response body, we'll return the status.
	if sdkErr != nil && httpResp != nil {
		if httpResp.Body != nil {
			defer httpResp.Body.Close()
			respBody, parseError := io.ReadAll(httpResp.Body)
			trimmedBody := strings.TrimSpace(string(respBody))
			if parseError == nil && len(trimmedBody) > 0 {
				return errors.New(trimmedBody)
			} else if httpResp.Status != "" {
				return errors.New(httpResp.Status)
			}
		}
	}
	// In case we can't parse the body, or if there's no body at all, return the original error.
	return sdkErr
}
