package testserver

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func (c *CloudRouter) HandleAllRolesRoute(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")

		switch namespace := r.URL.Query().Get("namespace"); namespace {
		case "ksql":
			_, err := io.WriteString(w, v2RbacKsqlRoles)
			require.NoError(t, err)
		case "datagovernance":
			_, err := io.WriteString(w, v2RbacSRRoles)
			require.NoError(t, err)
		default:
			var allRoles []string = findAllPublicRolesSorted(v2RbacRoles)
			var response = "[" + strings.Join(allRoles, ",") + "]"
			_, err := io.WriteString(w, response)
			require.NoError(t, err)
		}
	}
}
