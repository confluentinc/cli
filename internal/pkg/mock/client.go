package mock

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	mdsmock "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

func NewClientMock() *ccloud.Client {
	return &ccloud.Client{
		Params:         nil,
		Auth:           &mock.Auth{},
		Account:        &mock.Account{},
		Kafka:          &mock.Kafka{},
		SchemaRegistry: &mock.SchemaRegistry{},
		Connect:        &mock.Connect{},
		User:           &mock.User{},
		APIKey:         &mock.APIKey{},
		KSQL:           &mock.KSQL{},
		MetricsApi:     &mock.MetricsApi{},
		UsageLimits:    &mock.UsageLimits{},
	}
}

func NewV2ClientMock() *ccloudv2.Client {
	return ccloudv2.NewClient(newIamClientMock(), "auth-token")
}

func newIamClientMock() *iamv2.APIClient {
	return &iamv2.APIClient{
		ServiceAccountsIamV2Api: &iammock.ServiceAccountsIamV2Api{},
		UsersIamV2Api:           &iammock.UsersIamV2Api{},
	}
}

func NewMdsClientMock() *mdsv2.APIClient {
	return &mdsv2.APIClient{RoleBindingsIamV2Api: &mdsmock.RoleBindingsIamV2Api{}}
}
