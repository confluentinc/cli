package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for: "/byok/v1/keys/{id}"
func handleByokKey(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokStoreV1 := fillByokStoreV1()
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		keyStr := vars["id"]
		switch r.Method {
		case http.MethodGet:
			handleByokKeyGet(t, keyStr, byokStoreV1)(w, r)
		case http.MethodDelete:
			handleByokKeyDelete(t, keyStr, byokStoreV1)(w, r)
		}
	}
}

func handleByokKeyGet(t *testing.T, keyStr string, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if keyStr == "UNKNOWN" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		byokKey := byokStoreV1[keyStr]
		err := json.NewEncoder(w).Encode(byokKey)
		require.NoError(t, err)
	}
}

func handleByokKeyDelete(t *testing.T, keyStr string, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if keystr exists in store
		if _, ok := byokStoreV1[keyStr]; !ok {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		delete(byokStoreV1, keyStr)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/byok/v1/keys/"
func handleByokKeys(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokStoreV1 := fillByokStoreV1()
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			handleByokKeysCreate(t, byokStoreV1)(w, r)
		} else if r.Method == http.MethodGet {
			handleByokKeysList(t, byokStoreV1)(w, r)
		}
	}
}

func handleByokKeysCreate(t *testing.T, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(byokv1.ByokV1Key)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		byokKey := new(byokv1.ByokV1Key)
		byokKey.Id = byokv1.PtrString(fmt.Sprintf("cck-%03d", 4))
		byokKey.Metadata = &byokv1.ObjectMeta{
			CreatedAt: byokv1.PtrTime(time.Date(2022, time.December, 24, 0, 0, 0, 0, time.UTC)),
		}
		byokKey.State = byokv1.PtrString("AVAILABLE")

		switch {
		case req.Key.ByokV1AwsKey != nil:
			byokKey.Key = &byokv1.ByokV1KeyKeyOneOf{
				ByokV1AwsKey: &byokv1.ByokV1AwsKey{
					KeyArn: req.Key.ByokV1AwsKey.KeyArn,
					Roles: &[]string{
						"arn:aws:iam::123456789012:role/role1",
						"arn:aws:iam::123456789012:role/role2",
					},
					Kind: req.Key.ByokV1AwsKey.Kind,
				},
			}
			byokKey.Provider = byokv1.PtrString("AWS")
		case req.Key.ByokV1AzureKey != nil:
			byokKey.Key = &byokv1.ByokV1KeyKeyOneOf{
				ByokV1AzureKey: &byokv1.ByokV1AzureKey{
					ApplicationId: byokv1.PtrString("12345678-1234-1234-1234-123456789012"),
					KeyId:         req.Key.ByokV1AzureKey.KeyId,
					KeyVaultId:    req.Key.ByokV1AzureKey.KeyVaultId,
					TenantId:      req.Key.ByokV1AzureKey.TenantId,
					Kind:          req.Key.ByokV1AzureKey.Kind,
				},
			}
			byokKey.Provider = byokv1.PtrString("AZURE")
		}

		byokStoreV1[*byokKey.Id] = byokKey
		err = json.NewEncoder(w).Encode(byokKey)
		require.NoError(t, err)
	}
}

func handleByokKeysList(t *testing.T, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokKeyList := byokKeysFilterV1(r.URL, byokStoreV1)
		err := json.NewEncoder(w).Encode(byokKeyList)
		require.NoError(t, err)
	}
}

func byokKeysFilterV1(url *url.URL, byokStoreV1 map[string]*byokv1.ByokV1Key) *byokv1.ByokV1KeyList {
	byokKeyList := byokv1.ByokV1KeyList{}
	q := url.Query()
	provider := q.Get("provider")
	state := q.Get("state")

	for _, key := range byokStoreV1 {
		providerFilter := provider == "" || provider == key.GetProvider()
		stateFilter := (state == "") || state == *key.State
		if providerFilter && stateFilter {
			byokKeyList.Data = append(byokKeyList.Data, *key)
		}
	}
	return &byokKeyList
}
