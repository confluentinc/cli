package flink

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	_nethttp "net/http"
	"os"
	"strings"

	"github.com/spf13/cobra"

	cmfsdk "github.com/confluentinc/cmf-sdk-go/v1"

	"github.com/confluentinc/cli/v4/pkg/auth"
	perrors "github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/utils"
	testserver "github.com/confluentinc/cli/v4/test/test-server"
)

type OnPremCMFRestFlagValues struct {
	url            string
	caCertPath     string
	clientCertPath string
	clientKeyPath  string
}

type CmfClientInterface interface {
	GetStatement(ctx context.Context, environment, name string) (cmfsdk.Statement, error)
	ListStatements(ctx context.Context, environment, computePool, status string) ([]cmfsdk.Statement, error)
	CreateStatement(ctx context.Context, environment string, statement cmfsdk.Statement) (cmfsdk.Statement, error)
	ListStatementExceptions(ctx context.Context, environment, statementName string) (cmfsdk.StatementExceptionList, error)
	DeleteStatement(ctx context.Context, environment, statement string) error
	UpdateStatement(ctx context.Context, environment, statementName string, statement cmfsdk.Statement) error
	GetStatementResults(ctx context.Context, environment, statementName, pageToken string) (cmfsdk.StatementResult, error)
	CmfApiContext() context.Context
}

type CmfRestClient struct {
	*cmfsdk.APIClient
	AuthToken string
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

func NewCmfRestClient(cfg *cmfsdk.Configuration, restFlags *OnPremCMFRestFlagValues, isTest bool) (*CmfRestClient, error) {
	var err error
	cmfRestClient := &CmfRestClient{}

	// Set server URL based on test or flag input
	if isTest {
		cfg.Servers = cmfsdk.ServerConfigurations{
			{
				URL:         testserver.TestCmfUrl.String(),
				Description: "Confluent Platform test CMF Server",
			},
		}
	} else {
		if restFlags.url == "" {
			return nil, perrors.NewErrorWithSuggestions(
				"url is required",
				"Specify a URL with `--url` or set the variable \"CONFLUENT_CMF_URL\".",
			)
		}

		cfg.Servers = cmfsdk.ServerConfigurations{
			{
				URL:         restFlags.url,
				Description: "Confluent Platform default CMF Server",
			},
		}
	}

	// Set the CMF specific HTTP client
	cfg.HTTPClient, err = NewCmfRestHttpClient(restFlags)
	if err != nil {
		return nil, err
	}

	// Build client
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

func (cmfClient *CmfRestClient) CmfApiContext() context.Context {
	if cmfClient.AuthToken == "" {
		return context.Background()
	}
	return context.WithValue(context.Background(), cmfsdk.ContextAccessToken, cmfClient.AuthToken)
}

// CreateApplication Create a Flink application in the specified environment.
// Internally, since the call for Create and Update is the same, we check if the environment doesn't contain said application before creation.
func (cmfClient *CmfRestClient) CreateApplication(ctx context.Context, environment string, application cmfsdk.FlinkApplication) (cmfsdk.FlinkApplication, error) {
	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, _ := cmfClient.FlinkApplicationsApi.GetApplication(ctx, environment, applicationName).Execute()
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusOK {
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`application "%s" already exists in the environment "%s"`, applicationName, environment)
	}

	outputApplication, httpResponse, err := cmfClient.FlinkApplicationsApi.CreateOrUpdateApplication(ctx, environment).FlinkApplication(application).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`failed to create application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}
	return outputApplication, nil
}

func (cmfClient *CmfRestClient) DeleteApplication(ctx context.Context, environment, application string) error {
	httpResp, err := cmfClient.FlinkApplicationsApi.DeleteApplication(ctx, environment, application).Execute()
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeApplication(ctx context.Context, environment, application string) (cmfsdk.FlinkApplication, error) {
	cmfApplication, httpResponse, err := cmfClient.FlinkApplicationsApi.GetApplication(ctx, environment, application).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`failed to describe application "%s" in the environment "%s": %s`, application, environment, parsedErr)
	}
	return cmfApplication, nil
}

func (cmfClient *CmfRestClient) ListApplications(ctx context.Context, environment string) ([]cmfsdk.FlinkApplication, error) {
	applications := make([]cmfsdk.FlinkApplication, 0)
	// 100 is an arbitrary page size we've chosen.
	var currentPageNumber int32 = 0
	const pageSize = 100
	done := false

	for !done {
		applicationsPage, httpResponse, err := cmfClient.FlinkApplicationsApi.GetApplications(ctx, environment).Page(currentPageNumber).Size(pageSize).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list applications in the environment "%s": %s`, environment, parsedErr)
		}
		applications = append(applications, applicationsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(applicationsPage.GetItems()), currentPageNumber)
	}

	return applications, nil
}

// UpdateApplication Update an application in the specified environment.
// Internally, since the call for Create and Update is the same, we check if the environment contains said application before updation.
func (cmfClient *CmfRestClient) UpdateApplication(ctx context.Context, environment string, application cmfsdk.FlinkApplication) (cmfsdk.FlinkApplication, error) {
	// Get the name of the application
	applicationName := application.Metadata["name"].(string)
	_, httpResponse, err := cmfClient.FlinkApplicationsApi.GetApplication(ctx, environment, applicationName).Execute()
	// check if the application exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`application "%s" does not exist in the environment "%s"`, applicationName, environment)
	} else if httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		// Any failure other than 404 is an error in the response and shouldn't be treated as the application not existing.
		parsedErr := parseSdkError(httpResponse, err)
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`failed to update application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}

	outputApplication, httpResponse, err := cmfClient.FlinkApplicationsApi.CreateOrUpdateApplication(ctx, environment).FlinkApplication(application).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.FlinkApplication{}, fmt.Errorf(`failed to update application "%s" in the environment "%s": %s`, applicationName, environment, parsedErr)
	}
	return outputApplication, nil
}

// CreateEnvironment Create an environment.
// Internally, since the call for Create and Update is the same, we check if the environment exists before creation.
func (cmfClient *CmfRestClient) CreateEnvironment(ctx context.Context, postEnvironment cmfsdk.PostEnvironment) (cmfsdk.Environment, error) {
	environmentName := postEnvironment.Name
	_, httpResponse, _ := cmfClient.EnvironmentsApi.GetEnvironment(ctx, environmentName).Execute()
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusOK {
		return cmfsdk.Environment{}, fmt.Errorf(`environment "%s" already exists`, environmentName)
	}

	outputEnvironment, httpResponse, err := cmfClient.EnvironmentsApi.CreateOrUpdateEnvironment(ctx).PostEnvironment(postEnvironment).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to create environment "%s": %s`, environmentName, parsedErr)
	}
	return outputEnvironment, nil
}

func (cmfClient *CmfRestClient) DeleteEnvironment(ctx context.Context, environment string) error {
	httpResp, err := cmfClient.EnvironmentsApi.DeleteEnvironment(ctx, environment).Execute()
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeEnvironment(ctx context.Context, environment string) (cmfsdk.Environment, error) {
	cmfEnvironment, httpResponse, err := cmfClient.EnvironmentsApi.GetEnvironment(ctx, environment).Execute()

	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to describe environment "%s": %s`, environment, parsedErr)
	}

	return cmfEnvironment, nil
}

// ListEnvironments Run through all the pages until we get an empty page, in that case, return.
func (cmfClient *CmfRestClient) ListEnvironments(ctx context.Context) ([]cmfsdk.Environment, error) {
	environments := make([]cmfsdk.Environment, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0

	for !done {
		environmentsPage, httpResponse, err := cmfClient.EnvironmentsApi.GetEnvironments(ctx).Page(currentPageNumber).Size(pageSize).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf("failed to list environments: %s", parsedErr)
		}

		environments = append(environments, environmentsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(environmentsPage.GetItems()), currentPageNumber)
	}

	return environments, nil
}

// UpdateEnvironment Create an environment.
// Internally, since the call for Create and Update is the same, we check if the environment exists before updation.
func (cmfClient *CmfRestClient) UpdateEnvironment(ctx context.Context, postEnvironment cmfsdk.PostEnvironment) (cmfsdk.Environment, error) {
	environmentName := postEnvironment.Name
	_, httpResponse, err := cmfClient.EnvironmentsApi.GetEnvironment(ctx, environmentName).Execute()
	// check if the environment exists by checking the status code
	if httpResponse != nil && httpResponse.StatusCode == http.StatusNotFound {
		return cmfsdk.Environment{}, fmt.Errorf(`environment "%s" does not exist`, environmentName)
	} else if httpResponse == nil || httpResponse.StatusCode != http.StatusOK {
		// Any failure other than 404 is an error in the response and shouldn't be treated as the environment not existing.
		parsedErr := parseSdkError(httpResponse, err)
		return cmfsdk.Environment{}, fmt.Errorf(`failed to update environment "%s": %s`, environmentName, parsedErr)
	}

	outputEnvironment, httpResponse, err := cmfClient.EnvironmentsApi.CreateOrUpdateEnvironment(ctx).PostEnvironment(postEnvironment).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Environment{}, fmt.Errorf(`failed to update environment "%s": %s`, environmentName, parsedErr)
	}
	return outputEnvironment, nil
}

func (cmfClient *CmfRestClient) CreateSavepointApplication(ctx context.Context, savepoint cmfsdk.Savepoint, environment, application string) (cmfsdk.Savepoint, error) {
	outputSavepoint, httpResponse, err := cmfClient.SavepointsApi.CreateSavepointForFlinkApplication(ctx, environment, application).Savepoint(savepoint).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Savepoint{}, fmt.Errorf(`failed to create savepoint "%s" in the environment "%s": %s`, savepoint.Metadata.GetName(), environment, parsedErr)
	}
	return outputSavepoint, nil
}

func (cmfClient *CmfRestClient) CreateSavepointStatement(ctx context.Context, savepoint cmfsdk.Savepoint, environment, statement string) (cmfsdk.Savepoint, error) {
	outputSavepoint, httpResponse, err := cmfClient.SavepointsApi.CreateSavepointForFlinkStatement(ctx, environment, statement).Savepoint(savepoint).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Savepoint{}, fmt.Errorf(`failed to create savepoint "%s" in the environment "%s": %s`, savepoint.Metadata.GetName(), environment, parsedErr)
	}
	return outputSavepoint, nil
}

func (cmfClient *CmfRestClient) DescribeSavepoint(ctx context.Context, environment, name, application, statement string) (cmfsdk.Savepoint, error) {
	var cmfSavepoint cmfsdk.Savepoint
	var httpResponse *_nethttp.Response
	var err error
	if statement != "" {
		cmfSavepoint, httpResponse, err = cmfClient.SavepointsApi.GetSavepointForFlinkStatement(ctx, environment, statement, name).Execute()
	} else {
		cmfSavepoint, httpResponse, err = cmfClient.SavepointsApi.GetSavepointForFlinkApplication(ctx, environment, application, name).Execute()
	}
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Savepoint{}, fmt.Errorf(`failed to describe savepoint "%s" in the environment "%s": %s`, name, environment, parsedErr)
	}
	return cmfSavepoint, nil
}

func (cmfClient *CmfRestClient) DetachSavepointApplication(ctx context.Context, savepoint, environment, application string) (cmfsdk.Savepoint, error) {
	outputSavepoint, httpResponse, err := cmfClient.SavepointsApi.DetachSavepointFromFlinkApplication(ctx, environment, application, savepoint).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Savepoint{}, fmt.Errorf(`failed to create savepoint in the environment "%s": %s`, environment, parsedErr)
	}
	return outputSavepoint, nil
}

func (cmfClient *CmfRestClient) DeleteSavepoint(ctx context.Context, environment, savepoint, application, statement string, force bool) error {
	if statement != "" {
		httpResp, err := cmfClient.SavepointsApi.DeleteSavepointForFlinkStatement(ctx, environment, statement, savepoint).Force(force).Execute()
		return parseSdkError(httpResp, err)
	} else {
		httpResp, err := cmfClient.SavepointsApi.DeleteSavepointForFlinkApplication(ctx, environment, application, savepoint).Force(force).Execute()
		return parseSdkError(httpResp, err)
	}
}

func (cmfClient *CmfRestClient) ListSavepoint(ctx context.Context, environment, statement, application string, isStatement bool) ([]cmfsdk.Savepoint, error) {
	savepoints := make([]cmfsdk.Savepoint, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0
	for !done {
		var savepointsPage cmfsdk.SavepointsPage
		var httpResponse *_nethttp.Response
		var err error
		if isStatement {
			savepointsPage, httpResponse, err = cmfClient.SavepointsApi.GetSavepointsForFlinkStatement(ctx, environment, statement).Page(currentPageNumber).Size(pageSize).Execute()
		} else {
			savepointsPage, httpResponse, err = cmfClient.SavepointsApi.GetSavepointsForFlinkApplication(ctx, environment, application).Page(currentPageNumber).Size(pageSize).Execute()
		}
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list savepoints in the environment "%s": %s`, environment, parsedErr)
		}
		savepoints = append(savepoints, savepointsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(savepointsPage.GetItems()), currentPageNumber)
	}
	return savepoints, nil
}

func (cmfClient *CmfRestClient) DescribeDetachedSavepoint(ctx context.Context, name string) (cmfsdk.Savepoint, error) {
	detachedSavepoint, httpResponse, err := cmfClient.DetachedSavepointsApi.GetDetachedSavepoint(ctx, name).Execute()

	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Savepoint{}, fmt.Errorf(`failed to describe detached savepoint "%s": %s`, name, parsedErr)
	}
	return detachedSavepoint, nil
}

func (cmfClient *CmfRestClient) ListDetachedSavepoint(ctx context.Context, filter string) ([]cmfsdk.Savepoint, error) {
	savepoints := make([]cmfsdk.Savepoint, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0

	for !done {
		savepointsPage, httpResponse, err := cmfClient.DetachedSavepointsApi.ListDetachedSavepoints(ctx).Page(currentPageNumber).Size(pageSize).Name(filter).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list detached savepoints %s`, parsedErr)
		}
		savepoints = append(savepoints, savepointsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(savepointsPage.GetItems()), currentPageNumber)
	}

	return savepoints, nil
}

func (cmfClient *CmfRestClient) DeleteDetachedSavepoint(ctx context.Context, name string) error {
	httpResp, err := cmfClient.DetachedSavepointsApi.DeleteDetachedSavepoint(ctx, name).Execute()
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) CreateComputePool(ctx context.Context, environment string, computePool cmfsdk.ComputePool) (cmfsdk.ComputePool, error) {
	computePoolName := computePool.Metadata.Name
	if computePoolName == "" {
		return cmfsdk.ComputePool{}, fmt.Errorf("compute pool name is required")
	}
	outputComputePool, httpResponse, err := cmfClient.SQLApi.CreateComputePool(ctx, environment).ComputePool(computePool).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.ComputePool{}, fmt.Errorf(`failed to create compute pool "%s" in the environment "%s": %s`, computePoolName, environment, parsedErr)
	}
	return outputComputePool, nil
}

func (cmfClient *CmfRestClient) DeleteComputePool(ctx context.Context, environment, computePool string) error {
	httpResp, err := cmfClient.SQLApi.DeleteComputePool(ctx, environment, computePool).Execute()
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) DescribeComputePool(ctx context.Context, environment, computePool string) (cmfsdk.ComputePool, error) {
	cmfComputePool, httpResponse, err := cmfClient.SQLApi.GetComputePool(ctx, environment, computePool).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.ComputePool{}, fmt.Errorf(`failed to describe compute pool "%s" in the environment "%s": %s`, computePool, environment, parsedErr)
	}
	return cmfComputePool, nil
}

func (cmfClient *CmfRestClient) ListComputePools(ctx context.Context, environment string) ([]cmfsdk.ComputePool, error) {
	computePools := make([]cmfsdk.ComputePool, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0

	for !done {
		computePoolsPage, httpResponse, err := cmfClient.SQLApi.GetComputePools(ctx, environment).Page(currentPageNumber).Size(pageSize).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list compute pools in the environment "%s": %s`, environment, parsedErr)
		}
		computePools = append(computePools, computePoolsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(computePoolsPage.GetItems()), currentPageNumber)
	}

	return computePools, nil
}

func (cmfClient *CmfRestClient) CreateStatement(ctx context.Context, environment string, statement cmfsdk.Statement) (cmfsdk.Statement, error) {
	statementName := statement.Metadata.Name
	outputStatement, httpResponse, err := cmfClient.SQLApi.CreateStatement(ctx, environment).Statement(statement).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Statement{}, fmt.Errorf(`failed to create Flink SQL statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return outputStatement, nil
}

func (cmfClient *CmfRestClient) GetStatement(ctx context.Context, environment, name string) (cmfsdk.Statement, error) {
	statement, httpResponse, err := cmfClient.SQLApi.GetStatement(ctx, environment, name).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.Statement{}, fmt.Errorf(`failed to get Flink SQL statement "%s" in the environment "%s": %s`, name, environment, parsedErr)
	}
	return statement, nil
}

func (cmfClient *CmfRestClient) UpdateStatement(ctx context.Context, environment, statementName string, statement cmfsdk.Statement) error {
	httpResponse, err := cmfClient.SQLApi.UpdateStatement(ctx, environment, statementName).Statement(statement).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return fmt.Errorf(`failed to update statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return nil
}

func (cmfClient *CmfRestClient) DeleteStatement(ctx context.Context, environment, statement string) error {
	httpResp, err := cmfClient.SQLApi.DeleteStatement(ctx, environment, statement).Execute()
	return parseSdkError(httpResp, err)
}

func (cmfClient *CmfRestClient) ListStatements(ctx context.Context, environment, computePool, status string) ([]cmfsdk.Statement, error) {
	statements := make([]cmfsdk.Statement, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0

	request := cmfClient.SQLApi.GetStatements(ctx, environment)
	if computePool != "" {
		request = request.ComputePool(computePool)
	}
	if status != "" {
		request = request.Phase(status)
	}

	for !done {
		statementsPage, httpResponse, err := request.Page(currentPageNumber).Size(pageSize).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list statements in the environment "%s": %s`, environment, parsedErr)
		}
		statements = append(statements, statementsPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(statementsPage.GetItems()), currentPageNumber)
	}

	return statements, nil
}

func (cmfClient *CmfRestClient) ListStatementExceptions(ctx context.Context, environment, statementName string) (cmfsdk.StatementExceptionList, error) {
	exceptionList, httpResponse, err := cmfClient.SQLApi.GetStatementExceptions(ctx, environment, statementName).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.StatementExceptionList{}, fmt.Errorf(`failed to list exceptions for statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return exceptionList, nil
}

func (cmfClient *CmfRestClient) GetStatementResults(ctx context.Context, environment, statementName, pageToken string) (cmfsdk.StatementResult, error) {
	req := cmfClient.SQLApi.GetStatementResult(ctx, environment, statementName)
	if pageToken != "" {
		req = req.PageToken(pageToken)
	}
	resp, httpResponse, err := req.Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.StatementResult{}, fmt.Errorf(`failed to get result for statement "%s" in the environment "%s": %s`, statementName, environment, parsedErr)
	}
	return resp, nil
}

func (cmfClient *CmfRestClient) CreateCatalog(ctx context.Context, kafkaCatalog cmfsdk.KafkaCatalog) (cmfsdk.KafkaCatalog, error) {
	catalogName := kafkaCatalog.Metadata.Name
	outputCatalog, httpResponse, err := cmfClient.SQLApi.CreateKafkaCatalog(ctx).KafkaCatalog(kafkaCatalog).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf(`failed to create Kafka Catalog "%s": %s`, catalogName, parsedErr)
	}
	return outputCatalog, nil
}

func (cmfClient *CmfRestClient) DescribeCatalog(ctx context.Context, catalogName string) (cmfsdk.KafkaCatalog, error) {
	outputCatalog, httpResponse, err := cmfClient.SQLApi.GetKafkaCatalog(ctx, catalogName).Execute()
	if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
		return cmfsdk.KafkaCatalog{}, fmt.Errorf(`failed to get Kafka Catalog "%s": %s`, catalogName, parsedErr)
	}
	return outputCatalog, nil
}

func (cmfClient *CmfRestClient) ListCatalog(ctx context.Context) ([]cmfsdk.KafkaCatalog, error) {
	catalogs := make([]cmfsdk.KafkaCatalog, 0)
	done := false
	// 100 is an arbitrary page size we've chosen.
	const pageSize = 100
	var currentPageNumber int32 = 0

	for !done {
		catalogPage, httpResponse, err := cmfClient.SQLApi.GetKafkaCatalogs(ctx).Page(currentPageNumber).Size(pageSize).Execute()
		if parsedErr := parseSdkError(httpResponse, err); parsedErr != nil {
			return nil, fmt.Errorf(`failed to list Kafka Catalog: %s`, parsedErr)
		}
		catalogs = append(catalogs, catalogPage.GetItems()...)
		currentPageNumber, done = extractPageOptions(len(catalogPage.GetItems()), currentPageNumber)
	}

	return catalogs, nil
}

func (cmfClient *CmfRestClient) DeleteCatalog(ctx context.Context, catalogName string) error {
	httpResp, err := cmfClient.SQLApi.DeleteKafkaCatalog(ctx, catalogName).Execute()
	return parseSdkError(httpResp, err)
}

// Returns the next page number and whether we need to fetch more pages or not.
func extractPageOptions(receivedItemsLength int, currentPageNumber int32) (int32, bool) {
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
