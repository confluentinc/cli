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
	OrgEnvironments = []*orgv2.OrgV2Environment{{Id: orgv2.PtrString("a-595"), DisplayName: orgv2.PtrString("default")},
		{Id: orgv2.PtrString("not-595"), DisplayName: orgv2.PtrString("other")},
		{Id: orgv2.PtrString("env-123"), DisplayName: orgv2.PtrString("env123")}, {Id: orgv2.PtrString(SRApiEnvId), DisplayName: orgv2.PtrString("srUpdate")}}
)

// Handler for: "/org/v2/environments/{id}"
func HandleOrgEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		w.Header().Set("Content-Type", "application/json")
		if env := isValidOrgEnvironmentId(OrgEnvironments, envId); env != nil {
			switch r.Method {
			case http.MethodDelete:
				_, err := io.WriteString(w, "")
				require.NoError(t, err)
			case http.MethodPatch: // `environment update {id} --name`
				req := orgv2.OrgV2Environment{}
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				env.DisplayName = req.DisplayName
			}
		} else {
			// env not found
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// Handler for: "/org/v2/environments"
func HandleOrgEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			environmentList := &orgv2.OrgV2EnvironmentList{Data: getOrgEnvironmentsList(OrgEnvironments)}
			err := json.NewEncoder(w).Encode(environmentList)
			require.NoError(t, err)
		}
	}
}

func getOrgEnvironmentsList(envs []*orgv2.OrgV2Environment) []orgv2.OrgV2Environment {
	envList := []orgv2.OrgV2Environment{}
	for _, env := range envs {
		envList = append(envList, *env)
	}
	return envList
}
