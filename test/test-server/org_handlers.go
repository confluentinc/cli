package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var (
	OrgEnvironments = []orgv2.OrgV2Environment{{Id: orgv2.PtrString("a-595"), DisplayName: orgv2.PtrString("default")},
		{Id: orgv2.PtrString("not-595"), DisplayName: orgv2.PtrString("other")},
		{Id: orgv2.PtrString("env-123"), DisplayName: orgv2.PtrString("env123")}, {Id: orgv2.PtrString(SRApiEnvId), DisplayName: orgv2.PtrString("srUpdate")}}
)

// Handler for: "/org/v2/environments/{id}"
func (c *V2Router) HandleOrgEnvironment(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		w.Header().Set("Content-Type", "application/json")
		if valid, env := isValidOrgEnvironmentId(OrgEnvironments, envId); valid {
			switch r.Method {
			case http.MethodGet:
				b, err := json.Marshal(env)
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			case http.MethodDelete:
				b, err := json.Marshal(http.Response{})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			}
		} else {
			// env not found
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// Handler for: "/org/v2/environments"
func (c *V2Router) HandleOrgEnvironments(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			environmentList := &orgv2.OrgV2EnvironmentList{Data: OrgEnvironments}
			b, err := json.Marshal(environmentList)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}
