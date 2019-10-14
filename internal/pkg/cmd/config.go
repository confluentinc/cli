package cmd

import (
	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

// KafkaCluster creates an KafkaV1 struct from the Kafka cluster of the current context.
func KafkaCluster(cfg *config.Config) (*kafkav1.KafkaCluster, error) {
	ctx := cfg.Context()
	if ctx == nil {
		return nil, errors.ErrNoContext
	}
	kcc, err := ctx.ActiveKafkaCluster()
	if err != nil {
		return nil, err
	}
	state, err := cfg.AuthenticatedState()
	if err != nil {
		return nil, err
	}
	return &kafkav1.KafkaCluster{AccountId: state.Auth.Account.Id, Id: kcc.ID, ApiEndpoint: kcc.APIEndpoint}, nil
}
