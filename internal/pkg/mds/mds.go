package mds

import (
	"context"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
)

func mdsApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), mdsv2.ContextAccessToken, authToken)
	return auth
}

func CreateIamRoleBinding(client mdsv2.APIClient, roleBinding mdsv2.IamV2RoleBinding, authToken string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	return client.RoleBindingsIamV2Api.CreateIamV2RoleBinding(mdsApiContext(authToken)).IamV2RoleBinding(roleBinding).Execute()
}

func DeleteIamRoleBinding(client mdsv2.APIClient, id, authToken string) (*http.Response, error) {
	return client.RoleBindingsIamV2Api.DeleteIamV2RoleBinding(mdsApiContext(authToken), id).Execute()
}

func GetIamRoleBinding(client mdsv2.APIClient, id, authToken string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	return client.RoleBindingsIamV2Api.GetIamV2RoleBinding(mdsApiContext(authToken), id).Execute()
}

func ListIamRoleBinding(client mdsv2.APIClient, principal, roleName, authToken string) ([]mdsv2.IamV2RoleBinding, *http.Response, error) {
	resp, r, err := client.RoleBindingsIamV2Api.ListIamV2RoleBindings(context.Background()).Principal(principal).RoleName(roleName).Execute()
	if err != nil {
		return nil, nil, err
	}
	return resp.Data, r, err
}
