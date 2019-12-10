package mock

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"
	"github.com/confluentinc/mds-sdk-go"
	mdsmock "github.com/confluentinc/mds-sdk-go/mock"
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
		KSQL:           &mock.MockKSQL{},
		Metrics:        &mock.Metrics{},
	}
}

func NewMDSClientMock() *mds.APIClient {
	return &mds.APIClient{
		AuthorizationApi:        mdsmock.AuthorizationApi{},
		ClusterVisibilityApi:    mdsmock.ClusterVisibilityApi{},
		RoleDefinitionsApi:      mdsmock.RoleDefinitionsApi{},
		TokensAuthenticationApi: mdsmock.TokensAuthenticationApi{},
		UserAndRoleMgmtApi:      mdsmock.UserAndRoleMgmtApi{},
	}
}
