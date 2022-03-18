package test_server

import (
	"encoding/json"
	"io"
	"io/ioutil"
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
func (c *V2Router) HandleOrgEnvironment(t *testing.T) func(http.ResponseWriter, *http.Request) {
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
				requestBody, err := ioutil.ReadAll(r.Body)
				require.NoError(t, err)
				err = json.Unmarshal(requestBody, &req)
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
func (c *V2Router) HandleOrgEnvironments(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			environmentList := &orgv2.OrgV2EnvironmentList{Data: getOrgEnvironmentsList(OrgEnvironments)}
			b, err := json.Marshal(environmentList)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
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
