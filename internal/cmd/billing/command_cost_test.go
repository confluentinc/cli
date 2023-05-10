package billing

import (
	"context"
	billingv1 "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1"
	billingMock "github.com/confluentinc/ccloud-sdk-go-v2/billing/v1/mock"
	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	climock "github.com/confluentinc/cli/mock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"testing"
)

type CostTestSuite struct {
	suite.Suite
	conf            *v1.Config
	billingCostMock *billingMock.CostsBillingV1Api
}

var costList = billingv1.BillingV1CostList{
	Data: []billingv1.BillingV1Cost{
		{
			StartDate:   billingv1.PtrString("2023-01-01 00:00:00"),
			EndDate:     billingv1.PtrString("2023-01-02 00:00:00"),
			Granularity: billingv1.PtrString("DAILY"),
			LineType:    billingv1.PtrString("LINE_TYPE_1"),
			Product:     billingv1.PtrString("KAFKA"),
			Resource: &billingv1.BillingV1Resource{
				Id:          billingv1.PtrString("lkc-123"),
				DisplayName: billingv1.PtrString("kafka_1"),
			},
			NetworkAccessType: billingv1.PtrString("INTERNET"),
			Price:             billingv1.PtrFloat64(0.123),
			Unit:              billingv1.PtrString("GB"),
			Amount:            billingv1.PtrFloat64(50.0),
			OriginalAmount:    billingv1.PtrFloat64(50.0),
		},
	},
}

func (suite *CostTestSuite) SetupTest() {
	suite.conf = v1.AuthenticatedCloudConfigMock()
	suite.billingCostMock = &billingMock.CostsBillingV1Api{
		ListBillingV1CostsFunc: func(_ context.Context) billingv1.ApiListBillingV1CostsRequest {
			return billingv1.ApiListBillingV1CostsRequest{}
		},
		ListBillingV1CostsExecuteFunc: func(_ billingv1.ApiListBillingV1CostsRequest) (billingv1.BillingV1CostList, *http.Response, error) {
			return costList, nil, nil
		},
	}
}

func (suite *CostTestSuite) newCmd(conf *v1.Config) *cobra.Command {
	billingClient := &billingv1.APIClient{
		CostsBillingV1Api: suite.billingCostMock,
	}
	prerunner := climock.NewPreRunnerMock(nil, &ccloudv2.Client{BillingClient: billingClient, AuthToken: "auth-token"}, nil, nil, conf)
	return newCostCommand(prerunner)
}

func (suite *CostTestSuite) TestListCosts() {
	suite.T().Run("valid arguments", func(t *testing.T) {

		cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
		cmd.SetArgs([]string{"list", "2021-01-01", "2021-02-01"})
		err := cmd.Execute()
		req := require.New(suite.T())
		req.Nil(err)
		req.True(suite.billingCostMock.ListBillingV1CostsCalled())
		req.True(suite.billingCostMock.ListBillingV1CostsExecuteCalled())
		suite.billingCostMock.Reset()
	})

	suite.T().Run("missing argument", func(t *testing.T) {
		cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
		cmd.SetArgs([]string{"list", "2021-01-01"})
		err := cmd.Execute()
		req := require.New(suite.T())
		req.Error(err)
		req.False(suite.billingCostMock.ListBillingV1CostsCalled())
		req.False(suite.billingCostMock.ListBillingV1CostsExecuteCalled())
		suite.billingCostMock.Reset()
	})

	suite.T().Run("wrong arg format", func(t *testing.T) {
		cmd := suite.newCmd(v1.AuthenticatedCloudConfigMock())
		cmd.SetArgs([]string{"list", "20-01-01", "2020-0-01"})
		err := cmd.Execute()
		req := require.New(suite.T())
		req.Error(err)
		req.False(suite.billingCostMock.ListBillingV1CostsCalled())
		req.False(suite.billingCostMock.ListBillingV1CostsExecuteCalled())
		suite.billingCostMock.Reset()
	})
}

func TestServiceAccountTestSuite(t *testing.T) {
	suite.Run(t, new(CostTestSuite))
}
