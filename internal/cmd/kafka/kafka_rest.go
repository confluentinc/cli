package kafka

import (
	"fmt"
	"strconv"
	"strings"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/dghubble/sling"
)

const kafkaPort = "8090"

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

//todo: add csv
func bootstrapServersToRestURL(bootstrap string) (string, error) {
	bootstrapServers := strings.Split(bootstrap, ",")

	server := bootstrapServers[0]
	serverLength := len(server)
	fmt.Println("server: " + server)
	if serverLength > 5 && server[serverLength-5] == ':' && isNumeric(server[serverLength-4:]) {
		//TODO: change to https when config is fixed
		fmt.Println("http://" + server[:serverLength-4] + kafkaPort + "/kafka/v3")
		return "http://" + server[:serverLength-4] + kafkaPort + "/kafka/v3", nil
	} else {
		return "", errors.New("Invalid bootstrap server.")
	}
}

func getKafkaRestSetup(kafkaClusterConfig *v1.KafkaClusterConfig) (string, error) {
	kafkaRestURL, err := bootstrapServersToRestURL(kafkaClusterConfig.Bootstrap)
	if err != nil {
		return "", err
	}
	return kafkaRestURL, nil
}

func getAccessToken(authenticatedState *v2.ContextState, server string) (string, error) {

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
