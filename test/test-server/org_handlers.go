package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	orgv2 "github.com/confluentinc/ccloud-sdk-go-v2/org/v2"
)

var OrgEnvironments = []*orgv2.OrgV2Environment{
	{Id: orgv2.PtrString("a-595"), DisplayName: orgv2.PtrString("default")},
	{Id: orgv2.PtrString("not-595"), DisplayName: orgv2.PtrString("other")},
	{Id: orgv2.PtrString("env-123"), DisplayName: orgv2.PtrString("env123")},
	{Id: orgv2.PtrString(SRApiEnvId), DisplayName: orgv2.PtrString("srUpdate")},
}

// Handler for: "/org/v2/environments/{id}"
func handleOrgEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		envId := mux.Vars(r)["id"]
		if i := slices.IndexFunc(OrgEnvironments, func(env *orgv2.OrgV2Environment) bool { return env.GetId() == envId }); i != -1 {
			env := OrgEnvironments[i]
			switch r.Method {
			case http.MethodGet:
				err := json.NewEncoder(w).Encode(env)
				require.NoError(t, err)
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
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

// Handler for: "/org/v2/environments"
func handleOrgEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			environmentList := &orgv2.OrgV2EnvironmentList{Data: getV2List(OrgEnvironments)}
			err := json.NewEncoder(w).Encode(environmentList)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/org/v2/organizations/{id}"
func handleOrgOrganization(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		displayName := "default"
		switch r.Method {
		case http.MethodPatch:
			req := &orgv2.OrgV2Environment{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			displayName = req.GetDisplayName()
		}

		organization := &orgv2.OrgV2Organization{
			Id:          orgv2.PtrString(id),
			DisplayName: orgv2.PtrString(displayName),
		}
		err := json.NewEncoder(w).Encode(organization)
		require.NoError(t, err)
	}
}

// Handler for: "/org/v2/organizations"
func handleOrgOrganizations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			organizationList := &orgv2.OrgV2OrganizationList{Data: []orgv2.OrgV2Organization{
				{Id: orgv2.PtrString("abc-123"), DisplayName: orgv2.PtrString("org1")},
				{Id: orgv2.PtrString("abc-456"), DisplayName: orgv2.PtrString("org2")},
				{Id: orgv2.PtrString("abc-789"), DisplayName: orgv2.PtrString("org3")},
			}}
			err := json.NewEncoder(w).Encode(organizationList)
			require.NoError(t, err)
		}
	}
}
