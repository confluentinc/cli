package org

import (
	"context"
	"net/http"

	org "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

func OrgApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), org.ContextAccessToken, authToken)
	return auth
}

func ListEnvironments(client *org.APIClient, authToken string) (org.OrgV2EnvironmentList, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.ListOrgV2Environments(OrgApiContext(authToken)).Execute()
}
