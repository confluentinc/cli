package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/iam/v2/users/{id}"
func handleIamUser(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["id"]
		switch userId {
		case "u-1":
			w.Header().Set("Content-Type", "application/json")
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		default:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/users"
func handleIamUsers(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users := []iamv2.IamV2User{
				buildIamUser("bstrauch@confluent.io", "Brian Strauch", "u11"),
				buildIamUser("mtodzo@confluent.io", "Miles Todzo", "u-17"),
				buildIamUser("u-11aaa@confluent.io", "11 Aaa", "u-11aaa"),
				buildIamUser("u-22bbb@confluent.io", "22 Bbb", "u-22bbb"),
				buildIamUser("u-33ccc@confluent.io", "33 Ccc", "u-33ccc"),
				buildIamUser("mhe@confluent.io", "Muwei He", "u-44ddd"),
			}
			userId := r.URL.Query().Get("id")
			if userId != "" {
				if userId == deactivatedResourceID {
					users = []iamv2.IamV2User{}
				}
			}
			res := iamv2.IamV2UserList{
				Data: users,
			}
			email := r.URL.Query().Get("email")
			if email != "" {
				for _, u := range users {
					if *u.Email == email {
						res = iamv2.IamV2UserList{
							Data: []iamv2.IamV2User{u},
						}
						break
					}
				}
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/service_accounts/{id}"
func handleIamServiceAccount(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodPatch:
			var req iamv2.IamV2ServiceAccount
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamv2.IamV2ServiceAccount{Id: iamv2.PtrString(id), Description: req.Description}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/service_accounts"
func handleIamServiceAccounts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString(serviceAccountResourceID),
				DisplayName: iamv2.PtrString("service_account"),
				Description: iamv2.PtrString("at your service."),
			}
			err := json.NewEncoder(w).Encode(iamv2.IamV2ServiceAccountList{Data: []iamv2.IamV2ServiceAccount{serviceAccount}})
			require.NoError(t, err)
		case http.MethodPost:
			var req iamv2.IamV2ServiceAccount
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString("sa-55555"),
				DisplayName: req.DisplayName,
				Description: req.Description,
			}
			err = json.NewEncoder(w).Encode(serviceAccount)
			require.NoError(t, err)
		}
	}
}
