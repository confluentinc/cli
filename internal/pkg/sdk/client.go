package sdk

import (
	"github.com/confluentinc/ccloud-sdk-go"

	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/sdk/apikey"
	"github.com/confluentinc/cli/internal/pkg/sdk/environment"
	"github.com/confluentinc/cli/internal/pkg/sdk/kafka"
	"github.com/confluentinc/cli/internal/pkg/sdk/ksql"
	"github.com/confluentinc/cli/internal/pkg/sdk/user"
)

type Client struct {
	BaseClient *ccloud.Client
	Logger     *log.Logger
	APIKey     ccloud.APIKey
	Account    ccloud.Account
	Kafka      ccloud.Kafka
	KSQL       ccloud.KSQL
	User       ccloud.User
}

func NewClient(baseClient *ccloud.Client, logger *log.Logger) *Client {
	return &Client{
		BaseClient: baseClient,
		Logger:     logger,
		APIKey:     &apikey.APIKey{Client: baseClient, Logger: logger},
		Account:    &environment.Environment{Client: baseClient, Logger: logger},
		Kafka:      &kafka.Kafka{Client: baseClient, Logger: logger},
		KSQL:       &ksql.KSQL{Client: baseClient, Logger: logger},
		User:       &user.User{Client: baseClient, Logger: logger},
	}
}
