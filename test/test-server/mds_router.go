package testserver

import (
	"io"
	"net/http"
	"path"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/types"
)

var mdsRoutes = []route{
	{"/security/1.0/authenticate", handleAuthenticate},
	{"/security/1.0/registry/clusters", handleRegistryClusters},
}

type MdsRouter struct {
	*mux.Router
}

func NewMdsRouter(t *testing.T) *MdsRouter {
	router := &MdsRouter{mux.NewRouter()}
	router.Use(defaultHeaderMiddleware)

	for _, route := range mdsRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	router.addRoutesAndReplies(t, "/security/1.0/roles", v1RoutesAndReplies, v1RbacRoles)
	router.addDefaultHandler(t)
	router.Handle("/api/metadata/security/v2alpha1/roles/InvalidRole", http.NotFoundHandler())

	return router
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
			_, err := io.WriteString(w, s)
			require.NoError(t, err)
		})
	}
}

func findAllPublicRolesSorted(rbacRoles map[string]string) []string {
	roleNames := types.GetSortedKeys(rbacRoles)

	allRoles := make([]string, len(roleNames))
	for i, name := range roleNames {
		allRoles[i] = rbacRoles[name]
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
