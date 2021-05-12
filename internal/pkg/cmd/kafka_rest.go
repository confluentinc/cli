package cmd

import (
	"context"
	"strings"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/dghubble/sling"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

type KafkaREST struct {
	Client  *kafkarestv3.APIClient
	Context context.Context
}

func NewKafkaREST(client *kafkarestv3.APIClient, context context.Context) *KafkaREST {
	return &KafkaREST{
		Client:  client,
		Context: context,
	}
}

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func getBearerToken(authenticatedState *v2.ContextState, server string) (string, error) {
	bearerSessionToken := "Bearer " + authenticatedState.AuthToken
	accessTokenEndpoint := strings.Trim(server, "/") + "/api/access_tokens"

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	_, err := sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).Body(strings.NewReader("{}")).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}

	return responses.Token, nil
}
