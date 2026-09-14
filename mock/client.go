package mock

import (
	"context"
	"net/http"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"
	mdsmock "github.com/confluentinc/mds-sdk-go-public/mdsv1/mock"
)

func NewCCloudClientFactoryMock(ctrl *gomock.Controller, auth *ccloudv1mock.Auth, userInterface *ccloudv1mock.UserInterface, req *require.Assertions) *MockCCloudClientFactory {
	factory := NewMockCCloudClientFactory(ctrl)
	factory.EXPECT().AnonHTTPClientFactory(gomock.Any()).DoAndReturn(func(baseURL string) *ccloudv1.Client {
		req.Equal("https://confluent.cloud", baseURL)
		return &ccloudv1.Client{
			Params: &ccloudv1.Params{HttpClient: new(http.Client)},
			Auth:   auth,
			User:   userInterface,
		}
	}).AnyTimes()
	factory.EXPECT().JwtHTTPClientFactory(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(func(ctx context.Context, jwt, baseURL string) *ccloudv1.Client {
		return &ccloudv1.Client{
			Growth: &ccloudv1mock.Growth{
				GetFreeTrialInfoFunc: func(_ int32) ([]*ccloudv1.GrowthPromoCodeClaim, error) {
					return []*ccloudv1.GrowthPromoCodeClaim{}, nil
				},
			},
			Auth: auth,
			User: userInterface,
		}
	}).AnyTimes()
	return factory
}

func NewMdsClientMock(token string) *mdsv1.APIClient {
	mdsConfig := mdsv1.NewConfiguration()
	mdsClient := mdsv1.NewAPIClient(mdsConfig)
	mdsClient.TokensAndAuthenticationApi = &mdsmock.TokensAndAuthenticationApi{
		GetTokenFunc: func(_ context.Context) (mdsv1.AuthenticationResponse, *http.Response, error) {
			res := mdsv1.AuthenticationResponse{
				AuthToken: token,
				TokenType: "JWT",
				ExpiresIn: 100,
			}
			return res, nil, nil
		},
	}
	return mdsClient
}
