package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/byok/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var (
	byokStoreV1 = map[string]*byokv1.ByokV1Key{}
	byokTime    = byokv1.PtrTime(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
	byokIndex   = int32(4)
)

func init() {
	fillByokStoreV1()
}

// Handler for: "/byok/v1/keys/{id}"
func handleByokKey(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		keyStr := vars["id"]
		switch r.Method {
		case http.MethodGet:
			handleByokKeyGet(t, keyStr)(w, r)
		case http.MethodDelete:
			handleByokKeyDelete(t, keyStr)(w, r)
		}
	}
}

func handleByokKeyGet(t *testing.T, keyStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if keyStr == "UNKNOWN" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		byokkey := byokStoreV1[keyStr]
		err := json.NewEncoder(w).Encode(byokkey)
		require.NoError(t, err)
	}
}

func handleByokKeyDelete(t *testing.T, keyStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		delete(byokStoreV1, keyStr)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/byok/v1/keys/"
func handleByokKeys(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			handleByokKeysCreate(t)(w, r)
		} else if r.Method == http.MethodGet {
			handleByokKeysList(t)(w, r)
		}
	}
}

func handleByokKeysCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(byokv1.ByokV1Key)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		byokKey := new(byokv1.ByokV1Key)
		byokKey.Id = byokv1.PtrString(fmt.Sprintf("cck-%03d", byokIndex))
		byokIndex++
		byokKey.Metadata = &byokv1.ObjectMeta{
			CreatedAt: byokTime,
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

func handleByokKeysList(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokKeyList := byokv1.ByokV1KeyList{}
		for _, key := range byokStoreV1 {
			byokKeyList.Data = append(byokKeyList.Data, *key)
		}
		err := json.NewEncoder(w).Encode(byokKeyList)
		require.NoError(t, err)
	}
}
