package schemaregistry

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	srMock "github.com/confluentinc/schema-registry-sdk-go/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	cliMock "github.com/confluentinc/cli/mock"
)

type ConfigTestSuite struct {
	suite.Suite
	cfg          *v1.Config
	srClientMock *srsdk.APIClient
}

func (suite *ConfigTestSuite) SetupSuite() {
	suite.cfg = v1.AuthenticatedCloudConfigMock()
	suite.srClientMock = &srsdk.APIClient{
		DefaultApi: &srMock.DefaultApi{
			GetTopLevelConfigFunc: func(_ context.Context) (srsdk.Config, *http.Response, error) {
				return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
			},
			GetSubjectLevelConfigFunc: func(_ context.Context, subject string, opts *srsdk.GetSubjectLevelConfigOpts) (srsdk.Config, *http.Response, error) {
				switch subject {
				case "configured":
					return srsdk.Config{CompatibilityLevel: "FULL"}, nil, nil
				default:
					return srsdk.Config{}, &http.Response{StatusCode: http.StatusNotFound}, errors.New(fmt.Sprintf(errors.NoSubjectLevelConfigErrorMsg, subject))
				}
			},
		},
	}
}

func (suite *ConfigTestSuite) SetupTest() {
}

func (suite *ConfigTestSuite) newCmd() *cobra.Command {
	return New(suite.cfg, cliMock.NewPreRunnerMock(nil, nil, nil, nil, suite.cfg), suite.srClientMock)
}

func (suite *ConfigTestSuite) TestConfigDescribeGlobal() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"config", "describe"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
}

func (suite *ConfigTestSuite) TestConfigDescribeSubjectSuccess() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"config", "describe", "--subject", "configured"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Nil(err)
}

func (suite *ConfigTestSuite) TestConfigDescribeSubjectFail() {
	cmd := suite.newCmd()
	cmd.SetArgs([]string{"config", "describe", "--subject", "unconfigured"})
	err := cmd.Execute()
	req := require.New(suite.T())
	req.Error(err)
	req.Contains(err.Error(), "does not have subject-level compatibility configured")
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}
