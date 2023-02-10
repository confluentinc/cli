package testserver

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
	OrgOrganizations = []*orgv2.OrgV2Organization{{Id: orgv2.PtrString("abc-123"), DisplayName: orgv2.PtrString("org1")},
		{Id: orgv2.PtrString("abc-456"), DisplayName: orgv2.PtrString("org2")},
		{Id: orgv2.PtrString("abc-789"), DisplayName: orgv2.PtrString("org3")}}
)

// Handler for: "/org/v2/environments/{id}"
func handleOrgEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			environment := &orgv2.OrgV2Environment{
				Id:          orgv2.PtrString(envId),
				DisplayName: orgv2.PtrString("default"),
			}
			err := json.NewEncoder(w).Encode(environment)
			require.NoError(t, err)
			return
		}

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
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

// Handler for: "/org/v2/environments"
func handleOrgEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			environmentList := &orgv2.OrgV2EnvironmentList{Data: getOrgResourcesList[orgv2.OrgV2Environment](OrgEnvironments)}
			err := json.NewEncoder(w).Encode(environmentList)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/org/v2/organizations/{id}"
func handleOrgOrganization(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		orgId := vars["id"]
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			if orgId == "abc-789" {
				w.WriteHeader(http.StatusForbidden)
			} else {
				organization := &orgv2.OrgV2Organization{
					Id:          orgv2.PtrString(orgId),
					DisplayName: orgv2.PtrString("default"),
				}
				err := json.NewEncoder(w).Encode(organization)
				require.NoError(t, err)
			}
		}
	}
}

// Handler for: "/org/v2/organizations"
func handleOrgOrganizations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			organizationList := &orgv2.OrgV2OrganizationList{Data: getOrgResourcesList[orgv2.OrgV2Organization](OrgOrganizations)}
			err := json.NewEncoder(w).Encode(organizationList)
			require.NoError(t, err)
		}
	}
}

func getOrgResourcesList[OrgV2Resource orgv2.OrgV2Environment | orgv2.OrgV2Organization] (resources []*OrgV2Resource) []OrgV2Resource {
	resList := []OrgV2Resource{}
	for _, res := range resources {
		resList = append(resList, *res)
	}
	return resList
}