package mock

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/ccloud-sdk-go/mock"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/sdk"
)

func NewEmptyClientMock(logger *log.Logger) *sdk.Client {
	baseClient := &ccloud.Client{
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
	return sdk.NewClient(baseClient, logger)
}
