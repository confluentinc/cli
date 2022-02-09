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
	return client.ServiceAccountsIamV2Api.CreateIamV2ServiceAccount(iamApiContext(authToken)).IamV2ServiceAccount(serviceAccount).Execute()
}

func DeleteIamServiceAccount(client iamv2.APIClient, id, authToken string) (*http.Response, error) {
	return client.ServiceAccountsIamV2Api.DeleteIamV2ServiceAccount(iamApiContext(authToken), id).Execute()
}

func GetIamServiceAccount(client iamv2.APIClient, id, authToken string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	return client.ServiceAccountsIamV2Api.GetIamV2ServiceAccount(iamApiContext(authToken), id).Execute()
}

func ListIamServiceAccounts(client iamv2.APIClient, authToken string) ([]iamv2.IamV2ServiceAccount, *http.Response, error) {
	resp, r, err := client.ServiceAccountsIamV2Api.ListIamV2ServiceAccounts(iamApiContext(authToken)).Execute()
	if err != nil {
		return nil, nil, err
	}
	return resp.Data, r, err
}

func UpdateIamServiceAccount(client iamv2.APIClient, id string, update iamv2.IamV2ServiceAccountUpdate, authToken string) (iamv2.IamV2ServiceAccount, *http.Response, error) {
	return client.ServiceAccountsIamV2Api.UpdateIamV2ServiceAccount(iamApiContext(authToken), id).IamV2ServiceAccountUpdate(update).Execute()
}

func DeleteIamUser(client iamv2.APIClient, id, authToken string) (*http.Response, error) {
	return client.UsersIamV2Api.DeleteIamV2User(iamApiContext(authToken), id).Execute()
}

func GetIamUser(client iamv2.APIClient, id, authToken string) (iamv2.IamV2User, *http.Response, error) {
	return client.UsersIamV2Api.GetIamV2User(iamApiContext(authToken), id).Execute()
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

func ListIamUsers(client iamv2.APIClient, authToken string) ([]iamv2.IamV2User, *http.Response, error) {
	resp, r, err := client.UsersIamV2Api.ListIamV2Users(iamApiContext(authToken)).Execute()
	if err != nil {
		return nil, nil, err
	}
	return resp.Data, r, err
}

// func UpdateIamUser(client iam.APIClient, id string, update iam.IamV2UserUpdate, authToken string) (iam.IamV2User, *http.Response, error) {
// 	return client.UsersIamV2Api.UpdateIamV2User(iamApiContext(authToken), id).IamV2UserUpdate(update).Execute()
// }
