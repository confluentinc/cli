package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

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
		id := mux.Vars(r)["id"]

		switch r.Method {
		case http.MethodGet:
			if id == "env-dne" || id == "other" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			environment := &orgv2.OrgV2Environment{
				Id:          orgv2.PtrString(id),
				DisplayName: orgv2.PtrString("default"),
			}
			err := json.NewEncoder(w).Encode(environment)
			require.NoError(t, err)
		case http.MethodDelete:
			_, err := io.WriteString(w, "")
			require.NoError(t, err)
		case http.MethodPatch:
			req := &orgv2.OrgV2Environment{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			req.Id = orgv2.PtrString(id)

			err = json.NewEncoder(w).Encode(req)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/org/v2/environments"
func handleOrgEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			environmentList := &orgv2.OrgV2EnvironmentList{Data: getOrgEnvironmentsList(OrgEnvironments)}
			err := json.NewEncoder(w).Encode(environmentList)
			require.NoError(t, err)
		case http.MethodPost:
			req := &orgv2.OrgV2Environment{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			environment := &orgv2.OrgV2Environment{
				Id:          orgv2.PtrString("a-5555"),
				DisplayName: orgv2.PtrString(req.GetDisplayName()),
			}
			err = json.NewEncoder(w).Encode(environment)
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

func getOrgEnvironmentsList(envs []*orgv2.OrgV2Environment) []orgv2.OrgV2Environment {
	envList := []orgv2.OrgV2Environment{}
	for _, env := range envs {
		envList = append(envList, *env)
	}
	return envList
}
