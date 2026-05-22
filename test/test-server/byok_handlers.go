package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	byokv1 "github.com/confluentinc/ccloud-sdk-go-v2/byok/v1"
)

// Handler for: "/byok/v1/keys/{id}"
func handleByokKey(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokStoreV1 := fillByokStoreV1()
		vars := mux.Vars(r)
		keyStr := vars["id"]
		switch r.Method {
		case http.MethodGet:
			handleByokKeyGet(t, keyStr, byokStoreV1)(w, r)
		case http.MethodPatch:
			handleByokKeyUpdate(t, keyStr, byokStoreV1)(w, r)
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
		if byokKey, ok := byokStoreV1[keyStr]; !ok {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		} else {
			err := json.NewEncoder(w).Encode(byokKey)
			require.NoError(t, err)
		}
	}
}

func handleByokKeyUpdate(t *testing.T, keyStr string, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// check if keystr exists in store
		if existingKey, ok := byokStoreV1[keyStr]; !ok {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		} else {
			req := new(byokv1.ByokV1KeyUpdate)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			// Update the display name if provided
			if req.DisplayName != nil {
				existingKey.DisplayName = req.DisplayName
			}

			// Return the updated key
			err = json.NewEncoder(w).Encode(existingKey)
			require.NoError(t, err)
		}
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
		switch r.Method {
		case http.MethodPost:
			handleByokKeysCreate(t, byokStoreV1)(w, r)
		case http.MethodGet:
			handleByokKeysList(t, byokStoreV1)(w, r)
		}
	}
}

func handleByokKeysCreate(t *testing.T, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(byokv1.ByokV1Key)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		byokKey := &byokv1.ByokV1Key{
			Id:          byokv1.PtrString(fmt.Sprintf("cck-%03d", 4)),
			DisplayName: req.DisplayName,
			Metadata:    &byokv1.ObjectMeta{CreatedAt: byokv1.PtrTime(time.Date(2022, time.December, 24, 0, 0, 0, 0, time.UTC))},
			State:       byokv1.PtrString("AVAILABLE"),
			Validation: &byokv1.ByokV1KeyValidation{
				Phase: "INITIALIZING",
				Since: time.Date(2022, time.December, 24, 0, 0, 0, 0, time.UTC),
			},
		}

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
		case req.Key.ByokV1GcpKey != nil:
			byokKey.Key = &byokv1.ByokV1KeyKeyOneOf{
				ByokV1GcpKey: &byokv1.ByokV1GcpKey{
					KeyId: req.Key.ByokV1GcpKey.KeyId,
					Kind:  req.Key.ByokV1GcpKey.Kind,
				},
			}
			byokKey.Provider = byokv1.PtrString("GCP")
		}

		byokStoreV1[*byokKey.Id] = byokKey
		err = json.NewEncoder(w).Encode(byokKey)
		require.NoError(t, err)
	}
}

func handleByokKeysList(t *testing.T, byokStoreV1 map[string]*byokv1.ByokV1Key) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		byokKeyList := byokKeysFilterV1(r.URL, byokStoreV1)
		setPageToken(byokKeyList, &byokKeyList.Metadata, r.URL)
		err := json.NewEncoder(w).Encode(byokKeyList)
		require.NoError(t, err)
	}
}

func byokKeysFilterV1(url *url.URL, byokStoreV1 map[string]*byokv1.ByokV1Key) *byokv1.ByokV1KeyList {
	byokKeyList := byokv1.ByokV1KeyList{}
	q := url.Query()
	provider := q.Get("provider")
	state := q.Get("state")
	validationRegion := q.Get("validation_region")
	validationPhase := q.Get("validation_phase")
	displayName := q.Get("display_name")
	keyIdentifier := q.Get("key")

	for _, key := range byokStoreV1 {
		providerFilter := provider == "" || provider == key.GetProvider()
		stateFilter := (state == "") || state == *key.State
		displayNameFilter := displayName == "" || (key.DisplayName != nil && displayName == key.GetDisplayName())

		// For keyIdentifier filter, check the appropriate field based on provider (partial matching)
		keyIdentifierFilter := keyIdentifier == ""
		if keyIdentifier != "" && key.Key != nil {
			switch {
			case key.Key.ByokV1AwsKey != nil:
				keyIdentifierFilter = strings.Contains(key.Key.ByokV1AwsKey.GetKeyArn(), keyIdentifier)
			case key.Key.ByokV1AzureKey != nil:
				keyIdentifierFilter = strings.Contains(key.Key.ByokV1AzureKey.GetKeyId(), keyIdentifier)
			case key.Key.ByokV1GcpKey != nil:
				keyIdentifierFilter = strings.Contains(key.Key.ByokV1GcpKey.GetKeyId(), keyIdentifier)
			}
		}

		// For validation filters, check if validation exists and matches
		validationRegionFilter := validationRegion == ""
		validationPhaseFilter := validationPhase == ""
		if key.Validation != nil {
			if validationRegion != "" {
				validationRegionFilter = validationRegion == key.Validation.GetRegion()
			}
			if validationPhase != "" {
				validationPhaseFilter = validationPhase == key.Validation.GetPhase()
			}
		}

		if providerFilter && stateFilter && displayNameFilter && keyIdentifierFilter && validationRegionFilter && validationPhaseFilter {
			byokKeyList.Data = append(byokKeyList.Data, *key)
		}
	}

	sort.Slice(byokKeyList.Data, func(i, j int) bool {
		return byokKeyList.Data[i].GetMetadata().CreatedAt.After(*byokKeyList.Data[j].GetMetadata().CreatedAt)
	})

	return &byokKeyList
}
