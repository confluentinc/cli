package mock

import (
	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	ccloudv1mock "github.com/confluentinc/ccloud-sdk-go-v1-public/mock"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	cmkmock "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2/mock"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	iammock "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2/mock"
	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	orgmock "github.com/confluentinc/ccloud-sdk-go-v2/org/v2/mock"
	servicequotav1 "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1"
	quotasmock "github.com/confluentinc/ccloud-sdk-go-v2/service-quota/v1/mock"

	"github.com/confluentinc/cli/internal/pkg/ccloudv2"
)

func NewClientMock() *ccloudv1.Client {
	return &ccloudv1.Client{
		Billing:        &ccloudv1mock.Billing{},
		SchemaRegistry: &ccloudv1mock.SchemaRegistry{},
		User:           &ccloudv1mock.UserInterface{},
	}
}

func NewV2ClientMock() *ccloudv2.Client {
	return &ccloudv2.Client{
		AuthToken: "auth-token",

		CmkClient:          newCmkClientMock(),
		IamClient:          newIamClientMock(),
		OrgClient:          newOrgClientMock(),
		ServiceQuotaClient: newQuotasClientMock(),
	}
}

func newCmkClientMock() *cmkv2.APIClient {
	return &cmkv2.APIClient{ClustersCmkV2Api: &cmkmock.ClustersCmkV2Api{}}
}

func newIamClientMock() *iamv2.APIClient {
	return &iamv2.APIClient{
		ServiceAccountsIamV2Api: &iammock.ServiceAccountsIamV2Api{},
		UsersIamV2Api:           &iammock.UsersIamV2Api{},
	}
}

func newOrgClientMock() *orgv2.APIClient {
	return &orgv2.APIClient{EnvironmentsOrgV2Api: &orgmock.EnvironmentsOrgV2Api{}}
}

func newQuotasClientMock() *servicequotav1.APIClient {
	return &servicequotav1.APIClient{AppliedQuotasServiceQuotaV1Api: &quotasmock.AppliedQuotasServiceQuotaV1Api{}}
}
