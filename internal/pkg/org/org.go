package org

import (
	"context"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

func OrgApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), orgv2.ContextAccessToken, authToken)
	return auth
}

func ListEnvironments(client *orgv2.APIClient, authToken string) (orgv2.OrgV2EnvironmentList, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.ListOrgV2Environments(OrgApiContext(authToken)).Execute()
}
