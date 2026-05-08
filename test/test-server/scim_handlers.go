package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

// Mock SCIM tokens storage - shared mutable state across all tests
var scimTokenCounter = 3

func getInitialSCIMTokensV2() map[string]*orgv2.OrgV2ScimToken {
	return map[string]*orgv2.OrgV2ScimToken{
		"scim_token-12345": {
			ApiVersion:     orgv2.PtrString("org/v2"),
			Kind:           orgv2.PtrString("ScimToken"),
			Id:             orgv2.PtrString("scim_token-12345"),
			ConnectionName: orgv2.PtrString("op-12345"),
			Token:          orgv2.PtrString("scim_secret_token_value_12345"),
			CreatedAt:      orgv2.PtrTime(time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)),
			ExpiresAt:      orgv2.PtrTime(time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC)),
		},
		"scim_token-67890": {
			ApiVersion:     orgv2.PtrString("org/v2"),
			Kind:           orgv2.PtrString("ScimToken"),
			Id:             orgv2.PtrString("scim_token-67890"),
			ConnectionName: orgv2.PtrString("op-12345"),
			Token:          orgv2.PtrString("scim_secret_token_value_67890"),
			CreatedAt:      orgv2.PtrTime(time.Date(2025, 2, 1, 1, 0, 0, 0, time.UTC)),
			ExpiresAt:      orgv2.PtrTime(time.Date(2026, 2, 1, 1, 0, 0, 0, time.UTC)),
		},
	}
}

var scimTokensV2 = getInitialSCIMTokensV2()

// Handler for: POST "/org/v2/scim-tokens"
func handleCreateSCIMTokenV2(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var request orgv2.InlineObject
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		// Create new token
		newTokenId := "scim_token-" + string(rune('0'+scimTokenCounter))

		// Use fixed timestamps for test consistency, incrementing by 1 second for each new token
		now := time.Date(2026, 2, 12, 10, 57, 58, 0, time.UTC).Add(time.Duration(scimTokenCounter-3) * time.Second)
		scimTokenCounter++

		var expiresAt time.Time

		// Use ExpireDurationMins from request if provided, otherwise default to 6 months
		if request.ExpireDurationMins != nil && *request.ExpireDurationMins > 0 {
			expiresAt = now.Add(time.Duration(*request.ExpireDurationMins) * time.Minute)
		} else {
			expiresAt = now.AddDate(0, 6, 0) // Default 6 months
		}

		token := &orgv2.OrgV2ScimToken{
			ApiVersion:     orgv2.PtrString("org/v2"),
			Kind:           orgv2.PtrString("ScimToken"),
			Id:             orgv2.PtrString(newTokenId),
			ConnectionName: orgv2.PtrString("op-12345"),
			Token:          orgv2.PtrString("scim_secret_token_value_" + newTokenId),
			CreatedAt:      orgv2.PtrTime(now),
			ExpiresAt:      orgv2.PtrTime(expiresAt),
		}

		scimTokensV2[newTokenId] = token

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(token))
	}
}

// Handler for: GET "/org/v2/scim-tokens"
func handleListSCIMTokensV2(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var tokens []orgv2.OrgV2ScimToken
		for _, token := range scimTokensV2 {
			// Don't include the token value in list response
			tokenCopy := *token
			tokenCopy.Token = nil
			tokens = append(tokens, tokenCopy)
		}

		firstLink := *orgv2.NewNullableString(orgv2.PtrString("https://api.confluent.cloud/org/v2/scim-tokens"))
		lastLink := *orgv2.NewNullableString(orgv2.PtrString("https://api.confluent.cloud/org/v2/scim-tokens"))

		reply := orgv2.OrgV2ScimTokenList{
			ApiVersion: "org/v2",
			Kind:       "ScimTokenList",
			Metadata: orgv2.ListMeta{
				First: firstLink,
				Last:  lastLink,
			},
			Data: tokens,
		}

		w.Header().Set("Content-Type", "application/json")
		require.NoError(t, json.NewEncoder(w).Encode(reply))
	}
}

// Handler for: DELETE "/org/v2/scim-tokens/{id}"
func handleDeleteSCIMTokenV2(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenId := vars["id"]

		if _, exists := scimTokensV2[tokenId]; !exists {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			errorResponse := map[string]interface{}{
				"errors": []map[string]interface{}{
					{
						"status": "404",
						"detail": "SCIM token not found",
					},
				},
			}
			require.NoError(t, json.NewEncoder(w).Encode(errorResponse))
			return
		}

		delete(scimTokensV2, tokenId)

		w.WriteHeader(http.StatusNoContent)
	}
}

// Register SCIM routes
func RegisterSCIMRoutes(router *mux.Router, t *testing.T) {
	// SCIM token v2 routes
	router.HandleFunc("/org/v2/scim-tokens", handleCreateSCIMTokenV2(t)).Methods("POST")
	router.HandleFunc("/org/v2/scim-tokens", handleListSCIMTokensV2(t)).Methods("GET")
	router.HandleFunc("/org/v2/scim-tokens/{id}", handleDeleteSCIMTokenV2(t)).Methods("DELETE")
}
