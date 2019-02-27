package common

import (
	"fmt"

	"github.com/codyaray/go-editor"
	"github.com/spf13/cobra"

	kafkav1 "github.com/confluentinc/ccloudapis/kafka/v1"
	"github.com/confluentinc/cli/shared"
)

var messages = map[error]string{
	shared.ErrNoContext:      "You must login to access Confluent Cloud.",
	shared.ErrUnauthorized:   "You must login to access Confluent Cloud.",
	shared.ErrExpiredToken:   "Your access to Confluent Cloud has expired. Please login again.",
	shared.ErrIncorrectAuth:  "You have entered an incorrect username or password. Please try again.",
	shared.ErrMalformedToken: "Your auth token has been corrupted. Please login again.",
	shared.ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	shared.ErrNotFound:       "Kafka cluster not found.", // TODO: parametrize ErrNotFound for better error messaging
}

// HandleError provides standard error messaging for common errors.
func HandleError(err error, cmd *cobra.Command) error {
	// Give an indication of successful completion
	if err == nil {
		return nil
	}

	// Intercept errors to prevent usage from being printed.
	if msg, ok := messages[err]; ok {
		cmd.SilenceUsage = true
		return fmt.Errorf(msg)
	}

	switch err.(type) {
	case shared.NotAuthenticatedError:
		cmd.SilenceUsage = true
		return err
	case editor.ErrEditing:
		cmd.SilenceUsage = true
		return err
	case shared.KafkaError:
		cmd.SilenceUsage = true
		return err
	}

	return err
}

// Cluster returns the current cluster context
func Cluster(config *shared.Config) (*kafkav1.KafkaCluster, error) {
	ctx, err := config.Context()
	if err != nil {
		return nil, err
	}

	conf, err := config.KafkaClusterConfig()
	if err != nil {
		return nil, err
	}

	return &kafkav1.KafkaCluster{AccountId: config.Auth.Account.Id, Id: ctx.Kafka, ApiEndpoint: conf.APIEndpoint}, nil
}
