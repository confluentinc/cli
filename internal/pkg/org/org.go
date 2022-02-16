package org

import (
	"context"
	"net/http"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

func orgApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), orgv2.ContextAccessToken, authToken)
	return auth
}

func CreateOrgEnvironment(client *orgv2.APIClient, environment orgv2.OrgV2Environment, authToken string) (orgv2.OrgV2Environment, *http.Response, error) {
	req := client.EnvironmentsOrgV2Api.CreateOrgV2Environment(orgApiContext(authToken)).OrgV2Environment(environment)
	return client.EnvironmentsOrgV2Api.CreateOrgV2EnvironmentExecute(req)

}

func GetOrgEnvironment(client *orgv2.APIClient, envId string, authToken string) (orgv2.OrgV2Environment, *http.Response, error) {
	req := client.EnvironmentsOrgV2Api.GetOrgV2Environment(orgApiContext(authToken), envId)
	return client.EnvironmentsOrgV2Api.GetOrgV2EnvironmentExecute(req)
}

func ListEnvironments(client *orgv2.APIClient, authToken string) (orgv2.OrgV2EnvironmentList, *http.Response, error) {
	req := client.EnvironmentsOrgV2Api.ListOrgV2Environments(orgApiContext(authToken))
	return client.EnvironmentsOrgV2Api.ListOrgV2EnvironmentsExecute(req)
}

func UpdateOrgEnvironment(client *orgv2.APIClient, envId string, updateEnvironment orgv2.OrgV2Environment, authToken string) (orgv2.OrgV2Environment, *http.Response, error) {
	req := client.EnvironmentsOrgV2Api.UpdateOrgV2Environment(orgApiContext(authToken), envId).OrgV2Environment(updateEnvironment)
	return client.EnvironmentsOrgV2Api.UpdateOrgV2EnvironmentExecute(req)
}

func DeleteOrgEnvironment(client *orgv2.APIClient, envId string, authToken string) (*http.Response, error) {
	req := client.EnvironmentsOrgV2Api.DeleteOrgV2Environment(orgApiContext(authToken), envId)
	return client.EnvironmentsOrgV2Api.DeleteOrgV2EnvironmentExecute(req)
}
