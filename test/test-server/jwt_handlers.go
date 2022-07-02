package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func (c *CloudRouter) HandleJwtToken(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("oh cloud jwt being called.")
		type CreateOAuthTokenReply struct {
			Token string `json:"token"`
		}
		err := json.NewEncoder(w).Encode(&CreateOAuthTokenReply{Token: "OAUTH_TOKEN"})
		require.NoError(t, err)
	}
}
