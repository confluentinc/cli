package testserver

import (
	"io"
	"net/http"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

const (
	authenticate     = "/security/1.0/authenticate"
	registryClusters = "/security/1.0/registry/clusters"
	v1Base           = "/security/1.0/roles"
	v2Base           = "/api/metadata/security/v2alpha1/roles"
	v2InvalidRole    = "/api/metadata/security/v2alpha1/roles/InvalidRole"
)

type MdsRouter struct {
	*mux.Router
}

func NewMdsRouter(t *testing.T) *MdsRouter {
	router := &MdsRouter{mux.NewRouter()}
	router.buildMdsHandler(t)
	return router
}

func (m MdsRouter) buildMdsHandler(t *testing.T) {
	m.HandleFunc(authenticate, m.HandleAuthenticate(t))
	m.HandleFunc(registryClusters, m.HandleRegistryClusters(t))
	m.addRoutesAndReplies(t, v1Base, v1RoutesAndReplies, v1RbacRoles)
	m.addDefaultHandler(t)
	m.Handle(v2InvalidRole, http.NotFoundHandler())
}

func (m MdsRouter) addDefaultHandler(t *testing.T) {
	m.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to mds `+r.URL.Path+`"}}`)
		require.NoError(t, err)
	})
}

func (m MdsRouter) addRoutesAndReplies(t *testing.T, base string, routesAndReplies, rbacRoles map[string]string) {
	addRoles(base, routesAndReplies, rbacRoles)
	addAllPublicRoles(base, routesAndReplies, rbacRoles)
	for route, reply := range routesAndReplies {
		s := reply
		m.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			require.NoError(t, err)
		})
	}
}

func findAllPublicRolesSorted(rbacRoles map[string]string) []string {
	var roleNames []string
	for roleName := range rbacRoles {
		roleNames = append(roleNames, roleName)
	}

	sort.Strings(roleNames)

	var allRoles []string
	for _, name := range roleNames {
		allRoles = append(allRoles, rbacRoles[name])
	}

	return allRoles
}

func addRoles(base string, routesAndReplies, rbacRoles map[string]string) {
	for roleName, roleInfo := range rbacRoles {
		routesAndReplies[path.Join(base, roleName)] = roleInfo
	}
}

func addAllPublicRoles(base string, routesAndReplies, rbacRoles map[string]string) {
	allRoles := findAllPublicRolesSorted(rbacRoles)
	routesAndReplies[base] = "[" + strings.Join(allRoles, ",") + "]"
}
