package mock

import (
	"github.com/confluentinc/ccloud-sdk-go-v1"
	"github.com/confluentinc/ccloud-sdk-go-v1/mock"
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
	}
}
