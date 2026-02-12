package testserver

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	"github.com/confluentinc/cli/v4/pkg/ccstructs"
)

// Mock SCIM tokens storage - initialized fresh for each handler call
var scimTokenCounter = 3

func getInitialSCIMTokens() map[string]*flowv1.SCIMToken {
	return map[string]*flowv1.SCIMToken{
		"scim_token-12345": {
			Id:             "scim_token-12345",
			ConnectionName: "op-12345",
			Token:          "scim_secret_token_value_12345",
			CreatedAt:      &types.Timestamp{Seconds: time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC).Unix()},
			ExpiresAt:      &types.Timestamp{Seconds: time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC).Unix()},
		},
		"scim_token-67890": {
			Id:             "scim_token-67890",
			ConnectionName: "op-12345",
			Token:          "scim_secret_token_value_67890",
			CreatedAt:      &types.Timestamp{Seconds: time.Date(2025, 2, 1, 1, 0, 0, 0, time.UTC).Unix()},
			ExpiresAt:      &types.Timestamp{Seconds: time.Date(2026, 2, 1, 1, 0, 0, 0, time.UTC).Unix()},
		},
	}
}

var scimTokens = getInitialSCIMTokens()

// Handler for: POST "/api/organizations/{orgId}/sso/{connectionName}/scim/tokens"
func handleCreateSCIMToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionName := vars["connectionName"]

		// Decode request body
		var request flowv1.CreateSCIMTokenRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

		// Create new token
		newTokenId := "scim_token-" + string(rune('0'+scimTokenCounter))

		// Use fixed timestamps for test consistency, incrementing by 1 second for each new token
		now := time.Date(2026, 2, 12, 10, 57, 58, 0, time.UTC).Add(time.Duration(scimTokenCounter-3) * time.Second)
		scimTokenCounter++

		var expiresAt time.Time

		// Use ExpireDurationMins from request if provided, otherwise default to 6 months
		if request.ExpireDurationMins > 0 {
			expiresAt = now.Add(time.Duration(request.ExpireDurationMins) * time.Minute)
		} else {
			expiresAt = now.AddDate(0, 6, 0) // Default 6 months
		}

		token := &flowv1.SCIMToken{
			Id:             newTokenId,
			ConnectionName: connectionName,
			Token:          "scim_secret_token_value_" + newTokenId,
			CreatedAt:      &types.Timestamp{Seconds: now.Unix()},
			ExpiresAt:      &types.Timestamp{Seconds: expiresAt.Unix()},
		}

		scimTokens[newTokenId] = token

		reply := &flowv1.CreateSCIMTokenReply{
			ScimToken: token,
		}

		b, err := ccstructs.MarshalJSONToBytes(reply)
		require.NoError(t, err)
		_, err = w.Write(b)
		require.NoError(t, err)
	}
}

// Handler for: GET "/api/organizations/{orgId}/sso/{connectionName}/scim/tokens"
func handleListSCIMTokens(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		connectionName := vars["connectionName"]

		var tokens []*flowv1.SCIMToken
		for _, token := range scimTokens {
			if token.ConnectionName == connectionName {
				// Don't include the token value in list response
				tokenCopy := *token
				tokenCopy.Token = ""
				tokens = append(tokens, &tokenCopy)
			}
		}

		reply := &flowv1.ListSCIMTokensReply{
			ScimTokens: tokens,
		}

		b, err := ccstructs.MarshalJSONToBytes(reply)
		require.NoError(t, err)
		_, err = w.Write(b)
		require.NoError(t, err)
	}
}

// Handler for: DELETE "/api/organizations/{orgId}/sso/{connectionName}/scim/tokens/{tokenId}"
func handleDeleteSCIMToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		tokenId := vars["tokenId"]

		if _, exists := scimTokens[tokenId]; !exists {
			w.WriteHeader(http.StatusNotFound)
			reply := &flowv1.DeleteSCIMTokenReply{
				Error: &corev1.Error{
					Code:    http.StatusNotFound,
					Message: "SCIM token not found",
				},
			}
			b, err := ccstructs.MarshalJSONToBytes(reply)
			require.NoError(t, err)
			_, err = w.Write(b)
			require.NoError(t, err)
			return
		}

		delete(scimTokens, tokenId)

		reply := &flowv1.DeleteSCIMTokenReply{}
		b, err := ccstructs.MarshalJSONToBytes(reply)
		require.NoError(t, err)
		_, err = w.Write(b)
		require.NoError(t, err)
	}
}

// Register SCIM routes
func RegisterSCIMRoutes(router *mux.Router, t *testing.T) {
	// SCIM token routes
	router.HandleFunc("/api/organizations/{orgId}/sso/{connectionName}/scim/tokens", handleCreateSCIMToken(t)).Methods("POST")
	router.HandleFunc("/api/organizations/{orgId}/sso/{connectionName}/scim/tokens", handleListSCIMTokens(t)).Methods("GET")
	router.HandleFunc("/api/organizations/{orgId}/sso/{connectionName}/scim/tokens/{tokenId}", handleDeleteSCIMToken(t)).Methods("DELETE")
}
