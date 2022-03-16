package schemaregistry

import (
	"context"
	"net/http"
	"os"
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	cliMock "github.com/confluentinc/cli/mock"
)

type CompatibilityTestSuite struct {
	suite.Suite
	conf         *v1.Config
	srClientMock *srsdk.APIClient
}

func (suite *CompatibilityTestSuite) SetupSuite() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			TestCompatibilityBySubjectNameFunc: func(_ context.Context, subject, version string, body srsdk.RegisterSchemaRequest, opts *srsdk.TestCompatibilityBySubjectNameOpts) (srsdk.CompatibilityCheckResponse, *http.Response, error) {
				return srsdk.CompatibilityCheckResponse{IsCompatible: true}, nil, nil
			},
		},
	}
}

func (suite *CompatibilityTestSuite) TearDownTest() {
	os.RemoveAll("people.avsc")
}

func (suite *CompatibilityTestSuite) newCmd() *cobra.Command {
	os.Create("people.avsc")
	return New(suite.conf, cliMock.NewPreRunnerMock(nil, nil, nil, suite.conf), suite.srClientMock)
}

func (suite *CompatibilityTestSuite) TestValidateCompatibilityBySubjectName() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"compatibility", "validate", "--schema", "people.avsc", "--subject", "person", "--version", "latest"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
}

func TestCompatibilityTestSuite(t *testing.T) {
	suite.Run(t, new(CompatibilityTestSuite))
}
