package flink

import (
	"context"
	"errors"
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
	perrors "github.com/confluentinc/cli/v4/pkg/errors"
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
			return nil, perrors.NewErrorWithSuggestions("url is required", "Specify a URL with `--url` or set the variable \"CONFLUENT_CMF_URL\" in place of this flag.")
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

func (cmfClient *CmfRestClient) CreateComputePool(ctx context.Context, environment string, computePool cmfsdk.ComputePool) (cmfsdk.ComputePool, error) {
	computePoolName := computePool.Metadata.Name
	if computePoolName == "" {
		return cmfsdk.ComputePool{}, fmt.Errorf("compute pool name is required")
	}
	outputComputePool, httpResponse, err := cmfClient.DefaultApi.CreateComputePool(ctx, environment, computePool)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.ComputePool{}, fmt.Errorf(`failed to create compute pool "%s" in the environment "%s": %s`, computePoolName, environment, parsedErr)
	}
	return outputComputePool, nil
}

func (cmfClient *CmfRestClient) DeleteComputePool(ctx context.Context, environment, computePool string) error {
	httpResp, err := cmfClient.DefaultApi.DeleteComputePool(ctx, environment, computePool)
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeComputePool(ctx context.Context, environment, computePool string) (cmfsdk.ComputePool, error) {
	cmfComputePool, httpResponse, err := cmfClient.DefaultApi.GetComputePool(ctx, environment, computePool)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.ComputePool{}, fmt.Errorf(`failed to describe compute pool "%s" in the environment "%s": %s`, computePool, environment, parsedErr)
	}
	return cmfComputePool, nil
}

func (cmfClient *CmfRestClient) ListComputePools(ctx context.Context, environment string) ([]cmfsdk.ComputePool, error) {
	computePools := make([]cmfsdk.ComputePool, 0)
	currentPageNumber := 0
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100

	pagingOptions := &cmfsdk.GetComputePoolsOpts{
		Page: optional.NewInt32(int32(currentPageNumber)),
		Size: optional.NewInt32(pageSize),
	}

	for !done {
		computePoolsPage, httpResponse, err := cmfClient.DefaultApi.GetComputePools(ctx, environment, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list compute pools in the environment "%s": %s`, environment, parsedErr)
		}
		computePools = append(computePools, computePoolsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(computePoolsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return computePools, nil
}

func (cmfClient *CmfRestClient) CreateStatement(ctx context.Context, environment string, statement cmfsdk.Statement) (cmfsdk.Statement, error) {
	statementName := statement.Metadata.Name
	outputStatement, httpResponse, err := cmfClient.DefaultApi.CreateStatement(ctx, environment, statement)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Statement{}, fmt.Errorf(`failed to create Flink SQL statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return outputStatement, nil
}

func (cmfClient *CmfRestClient) GetStatement(ctx context.Context, environment, name string) (cmfsdk.Statement, error) {
	statement, httpResponse, err := cmfClient.DefaultApi.GetStatement(ctx, environment, name)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Statement{}, fmt.Errorf(`failed to get Flink SQL statement "%s" in the environment "%s": %s`, name, environment, parsedErr)
	}
	return statement, nil
}

func (cmfClient *CmfRestClient) UpdateStatement(ctx context.Context, environment, statementName string, statement cmfsdk.Statement) error {
	httpResponse, err := cmfClient.DefaultApi.UpdateStatement(ctx, environment, statementName, statement)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to update statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return nil
}

func (cmfClient *CmfRestClient) DeleteStatement(ctx context.Context, environment, statement string) error {
	httpResp, err := cmfClient.DefaultApi.DeleteStatement(ctx, environment, statement)
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) ListStatements(ctx context.Context, environment, computePool, status string) ([]cmfsdk.Statement, error) {
	statements := make([]cmfsdk.Statement, 0)
	currentPageNumber := 0
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100

	pagingOptions := &cmfsdk.GetStatementsOpts{
		Page: optional.NewInt32(int32(currentPageNumber)),
		Size: optional.NewInt32(pageSize),
	}
	if computePool != "" {
		pagingOptions.ComputePool = optional.NewString(computePool)
	}
	if status != "" {
		pagingOptions.Phase = optional.NewString(status)
	}

	for !done {
		statementsPage, httpResponse, err := cmfClient.DefaultApi.GetStatements(ctx, environment, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list statements in the environment "%s": %s`, environment, parsedErr)
		}
		statements = append(statements, statementsPage.Items...)
		currentPageNumber, done = extractPageOptions(len(statementsPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return statements, nil
}

// TODO: Check with Fabian to see if the pagination is needed for this exception list command
func (cmfClient *CmfRestClient) ListStatementExceptions(ctx context.Context, environment, statementName string) (cmfsdk.StatementExceptionList, error) {
	exceptionList, httpResponse, err := cmfClient.DefaultApi.GetStatementExceptions(ctx, environment, statementName)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.StatementExceptionList{}, fmt.Errorf(`failed to list exceptions for statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return exceptionList, nil
}

func (cmfClient *CmfRestClient) CreateCatalog(ctx context.Context, kafkaCatalog cmfsdk.KafkaCatalog) (cmfsdk.KafkaCatalog, error) {
	catalogName := kafkaCatalog.Metadata.Name
	outputCatalog, httpResponse, err := cmfClient.DefaultApi.CreateKafkaCatalog(ctx, kafkaCatalog)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf(`failed to create Kafka Catalog "%s": %s`, catalogName, parsedErr)
	}
	return outputCatalog, nil
}

func (cmfClient *CmfRestClient) DescribeCatalog(ctx context.Context, catalogName string) (cmfsdk.KafkaCatalog, error) {
	outputCatalog, httpResponse, err := cmfClient.DefaultApi.GetKafkaCatalog(ctx, catalogName)
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf(`failed to get Kafka Catalog "%s": %s`, catalogName, parsedErr)
	}
	return outputCatalog, nil
}

func (cmfClient *CmfRestClient) ListCatalog(ctx context.Context) ([]cmfsdk.KafkaCatalog, error) {
	catalogs := make([]cmfsdk.KafkaCatalog, 0)
	currentPageNumber := 0
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100

	pagingOptions := &cmfsdk.GetKafkaCatalogsOpts{
		Page: optional.NewInt32(int32(currentPageNumber)),
		Size: optional.NewInt32(pageSize),
	}

	for !done {
		catalogPage, httpResponse, err := cmfClient.DefaultApi.GetKafkaCatalogs(ctx, pagingOptions)
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list Kafka Catalog: %s`, parsedErr)
		}
		catalogs = append(catalogs, catalogPage.Items...)
		currentPageNumber, done = extractPageOptions(len(catalogPage.Items), currentPageNumber)
		pagingOptions.Page = optional.NewInt32(int32(currentPageNumber))
	}

	return catalogs, nil
}

func (cmfClient *CmfRestClient) DeleteCatalog(ctx context.Context, catalogName string) error {
	httpResp, err := cmfClient.DefaultApi.DeleteKafkaCatalog(ctx, catalogName)
	return parseSdkError(httpResp, err)
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
