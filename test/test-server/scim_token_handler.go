package testserver

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

// Handler for "/org/v2/scim-tokens"
func handleOrgV2ScimTokens(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			scimToken := readOrgV2ScimTokenFile(t, "read_created_scim_token.json")

			scimTokenList := &orgv2.OrgV2ScimTokenList{
				Data: []orgv2.OrgV2ScimToken{scimToken},
			}

			err := json.NewEncoder(w).Encode(scimTokenList)
			require.NoError(t, err)
		case http.MethodPost:
			scimToken := readOrgV2ScimTokenFile(t, "create_scim_token.json")

			// Overwrite updated fields using the request body
			err := json.NewDecoder(r.Body).Decode(&scimToken)
			require.NoError(t, err)

			err = json.NewEncoder(w).Encode(scimToken)
			require.NoError(t, err)
		}
	}
}

// Handler for "/org/v2/scim-tokens/{id}"
func handleOrgV2ScimTokensId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodDelete:
			switch id {
			case "invalid":
				w.WriteHeader(http.StatusNotFound)
			default:
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

func readOrgV2ScimTokenFile(t *testing.T, filename string) orgv2.OrgV2ScimToken {
	jsonPath := filepath.Join("test", "fixtures", "input", "org", "scim_token", filename)
	jsonData, err := os.ReadFile(jsonPath)
	require.NoError(t, err)

	scimToken := orgv2.OrgV2ScimToken{}
	err = json.Unmarshal(jsonData, &scimToken)
	require.NoError(t, err)

	return scimToken
}
