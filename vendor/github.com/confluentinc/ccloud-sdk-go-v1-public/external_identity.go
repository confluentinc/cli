package ccloud

import (
	"fmt"
	"net/http"

	"github.com/dghubble/sling"
)

var ErrFailedToCreateExternalIdentity = fmt.Errorf("failed to create external identity")

type ExternalIdentityService struct {
	base     *http.Client
	client   *http.Client
	sling    *sling.Sling
	apiSling *sling.Sling
	logger   Logger
}

func NewExternalIdentityService(client *Client) *ExternalIdentityService {
	return &ExternalIdentityService{
		base:     client.BaseClient,
		client:   client.HttpClient,
		sling:    client.sling.New().Add("Content-Type", "application/json"),
		apiSling: GetSlingWithNewClient(client.sling, client.BaseClient, client.Logger).ResponseDecoder(JsonDecoder{}),
		logger:   client.Logger,
	}
}

func (s *ExternalIdentityService) CreateExternalIdentity(cloud, accountID string) (string, error) {
	responseBody := new(CreateExternalIdentityResponse)
	response, err := s.sling.New().Post("/api/external_identities").BodyJSON(&CreateExternalIdentityRequest{
		Cloud:     cloud,
		AccountId: accountID,
	}).Receive(responseBody, nil)
	if err != nil {
		return "", E(ErrFailedToCreateExternalIdentity.Error())
	}
	if response.StatusCode >= 400 {
		return "", E(response.StatusCode, ErrFailedToCreateExternalIdentity.Error())
	}
	return responseBody.IdentityName, nil
}
