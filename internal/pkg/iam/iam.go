package iam

import (
	"context"
	"fmt"
	"net/http"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

func iamApiContext(authToken string) context.Context {
	auth := context.WithValue(context.Background(), iamv2.ContextAccessToken, authToken)
	return auth
}

func CreateIamServiceAccount(client iamv2.APIClient, serviceAccount iamv2.IamV2ServiceAccount, authToken string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := client.ServiceAccountsIamV2Api.CreateIamV2ServiceAccount(iamApiContext(authToken)).IamV2ServiceAccount(serviceAccount)
	return client.ServiceAccountsIamV2Api.CreateIamV2ServiceAccountExecute(req)
}

func DeleteIamServiceAccount(client iamv2.APIClient, id, authToken string) (*http.Response, error) {
	req := client.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccount(iamApiContext(authToken), id)
	return client.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccountExecute(req)
}

func GetIamServiceAccount(client iamv2.APIClient, id, authToken string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := client.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(iamApiContext(authToken), id)
	return client.ServiceAccountsIamV2Api.GetIamV2ServiceAccountExecute(req)
}

func ListIamServiceAccounts(client iamv2.APIClient, authToken string) (iamv2.IamV2ServiceAccountList, *http.Response, error) {
	req := client.ServiceAccountsIamV2Api.ListIamV2ServiceAccounts(iamApiContext(authToken))
	return client.ServiceAccountsIamV2Api.ListIamV2ServiceAccountsExecute(req)
}

func UpdateIamServiceAccount(client iamv2.APIClient, id string, update iamv2.IamV2ServiceAccountUpdate, authToken string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	req := client.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccount(iamApiContext(authToken), id).IamV2ServiceAccountUpdate(update)
	return client.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccountExecute(req)
}

func DeleteIamUser(client iamv2.APIClient, id, authToken string) (*http.Response, error) {
	req := client.UsersIamV2Api.DeleteIamV2User(iamApiContext(authToken), id)
	return client.UsersIamV2Api.DeleteIamV2UserExecute(req)
}

func GetIamUser(client iamv2.APIClient, id, authToken string) (iamv2.IamV2User, *http.Response, error) {
	req := client.UsersIamV2Api.GetIamV2User(iamApiContext(authToken), id)
	return client.UsersIamV2Api.GetIamV2UserExecute(req)
}

func GetIamUserByEmail(client iamv2.APIClient, email, authToken string) (iamv2.IamV2User, error) {
	resp, _, err := client.UsersIamV2Api.ListIamV2Users(iamApiContext(authToken)).Execute()
	if err != nil {
		return iamv2.IamV2User{}, err
	}
	users := resp.Data
	for _, user := range users {
		if email == *user.Email {
			return user, nil
		}
	}
	return iamv2.IamV2User{}, fmt.Errorf(errors.InvalidEmailErrorMsg, email)
}

func ListIamUsers(client iamv2.APIClient, authToken string) (iamv2.IamV2UserList, *http.Response, error) {
	req := client.UsersIamV2Api.ListIamV2Users(iamApiContext(authToken))
	return client.UsersIamV2Api.ListIamV2UsersExecute(req)
}

// func UpdateIamUser(client iam.APIClient, id string, update iam.IamV2UserUpdate, authToken string) (iam.IamV2User, *http.Response, error) {
// 	return client.UsersIamV2Api.UpdateIamV2User(iamApiContext(authToken), id).IamV2UserUpdate(update).Execute()
// }
