package cmd

import (
	"context"
	"strconv"
	"strings"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/dghubble/sling"

	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
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

const kafkaRestPort = "8090"

func bootstrapServersToRestURL(bootstrap string) (string, error) {
	bootstrapServers := strings.Split(bootstrap, ",")

	server := bootstrapServers[0]
	serverLength := len(server)

	if serverLength <= 6 {
		return "", errors.New(errors.InvalidBootstrapServerErrorMsg)
	}

	if _, err := strconv.Atoi(server[serverLength-4:]); err == nil && server[serverLength-5] == ':' {
		return "https://" + server[:serverLength-4] + kafkaRestPort + "/kafka/v3", nil
	}

	if _, err := strconv.Atoi(server[serverLength-5:]); err == nil && server[serverLength-6] == ':' {
		return "https://" + server[:serverLength-5] + kafkaRestPort + "/kafka/v3", nil
	}

	return "", errors.New(errors.InvalidBootstrapServerErrorMsg)
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
