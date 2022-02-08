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

func CreateOrgEnvironment(client *org.APIClient, environment org.OrgV2Environment, authToken string) (org.OrgV2Environment, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.CreateOrgV2Environment(OrgApiContext(authToken)).OrgV2Environment(environment).Execute()

}

func GetOrgEnvironment(client *org.APIClient, envId string, authToken string) (org.OrgV2Environment, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.GetOrgV2Environment(OrgApiContext(authToken), envId).Execute()
}

func ListEnvironments(client *org.APIClient, authToken string) (org.OrgV2EnvironmentList, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.ListOrgV2Environments(OrgApiContext(authToken)).Execute()
}

func UpdateOrgEnvironment(client *org.APIClient, envId string, updateEnvironment org.OrgV2Environment, authToken string) (org.OrgV2Environment, *http.Response, error) {
	return client.EnvironmentsOrgV2Api.UpdateOrgV2Environment(OrgApiContext(authToken), envId).OrgV2Environment(updateEnvironment).Execute()
}

func DeleteOrgEnvironment(client *org.APIClient, envId string, authToken string) (*http.Response, error) {
	return client.EnvironmentsOrgV2Api.DeleteOrgV2Environment(OrgApiContext(authToken), envId).Execute()
}
