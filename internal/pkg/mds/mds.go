package mds

import (
	"context"
	"net/http"

	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
)

func MdsApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), mdsv2.ContextAccessToken, authToken)
	return auth
}

func CreateIamRoleBinding(client mdsv2.APIClient, roleBinding mdsv2.IamV2RoleBinding, authToken string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := client.RoleBindingsIamV2Api.CreateIamV2RoleBinding(MdsApiContext(authToken)).IamV2RoleBinding(roleBinding)
	return client.RoleBindingsIamV2Api.CreateIamV2RoleBindingExecute(req)
}

func DeleteIamRoleBinding(client mdsv2.APIClient, id, authToken string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := client.RoleBindingsIamV2Api.DeleteIamV2RoleBinding(MdsApiContext(authToken), id)
	return client.RoleBindingsIamV2Api.DeleteIamV2RoleBindingExecute(req)
}

func GetIamRoleBinding(client mdsv2.APIClient, id, authToken string) (mdsv2.IamV2RoleBinding, *http.Response, error) {
	req := client.RoleBindingsIamV2Api.GetIamV2RoleBinding(MdsApiContext(authToken), id)
	return client.RoleBindingsIamV2Api.GetIamV2RoleBindingExecute(req)
}

func ListIamRoleBinding(client mdsv2.APIClient, principal, role, crnPattern, authToken string) (mdsv2.IamV2RoleBindingList, *http.Response, error) {
	req := client.RoleBindingsIamV2Api.ListIamV2RoleBindings(MdsApiContext(authToken))
	if principal != "" {
		req = req.Principal(principal)
	}
	if role != "" {
		req = req.RoleName(role)
	}
	if crnPattern != "" {
		req = req.CrnPattern(crnPattern)
	}
	return client.RoleBindingsIamV2Api.ListIamV2RoleBindingsExecute(req)
}
