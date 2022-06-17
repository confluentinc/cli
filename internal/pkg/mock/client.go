package mock

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	orgmock "github.com/confluentinc/ccloud-sdk-go-v2/org/v2/mock"
	quotasv2 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2"
	quotasmock "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v2/mock"

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
	return &ccloudv2.Client{
		AuthToken:          "auth-token",
		CmkClient:          newCmkClientMock(),
		IamClient:          newIamClientMock(),
		OrgClient:          newOrgClientMock(),
		ServiceQuotaClient: newQuotasClientMock(),
	}
}

func newCmkClientMock() *cmkv2.APIClient {
	return &cmkv2.APIClient{ClustersCmkV2Api: &cmkmock.ClustersCmkV2Api{}}
}

func newOrgClientMock() *orgv2.APIClient {
	return &orgv2.APIClient{EnvironmentsOrgV2Api: &orgmock.EnvironmentsOrgV2Api{}}
}

func newIamClientMock() *iamv2.APIClient {
	return &iamv2.APIClient{
		ServiceAccountsIamV2Api: &iammock.ServiceAccountsIamV2Api{},
		UsersIamV2Api:           &iammock.UsersIamV2Api{},
	}
}

func newQuotasClientMock() *quotasv2.APIClient {
	return &quotasv2.APIClient{AppliedQuotasServiceQuotaV2Api: &quotasmock.AppliedQuotasServiceQuotaV2Api{}}
}
