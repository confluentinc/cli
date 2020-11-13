package kafka

import (
	"strings"

	"github.com/dghubble/sling"
	"github.com/spf13/cobra"
)

const kafkaPort = "8090"

type response struct {
	Error string `json:"error"`
	Token string `json:"token"`
}

func getServerURL(bootstrap string) string {
	return "http://" + strings.TrimSuffix(bootstrap, "9092") + kafkaPort + "/kafka/v3"
}

func getKafkaRestSetup(a *authenticatedTopicCommand, cmd *cobra.Command) (string, string, error) {
	kcc, err := a.AuthenticatedCLICommand.Context.GetKafkaClusterForCommand(cmd)
	if err != nil {
		return "", "", err
	}
	lkc := kcc.ID
	bootstrap := kcc.Bootstrap
	kafkaRestURL := getServerURL(bootstrap)

	return lkc, kafkaRestURL, nil
}

func getAccessToken(a *authenticatedTopicCommand, cmd *cobra.Command) (string, error) {
	state, err := a.AuthenticatedCLICommand.Context.AuthenticatedState(cmd)
	if err != nil {
		return "", err
	}
	bearerSessionToken := "Bearer " + state.AuthToken
	accessTokenEndpoint := strings.Trim(a.Context.Platform.Server, "/") + "/api/access_tokens"

	// Configure and send post request with session token to Auth Service to get access token
	responses := new(response)
	_, err = sling.New().Add("content", "application/json").Add("Content-Type", "application/json").Add("Authorization", bearerSessionToken).Body(strings.NewReader("{}")).Post(accessTokenEndpoint).ReceiveSuccess(responses)
	if err != nil {
		return "", err
	}

	return responses.Token, nil
}
