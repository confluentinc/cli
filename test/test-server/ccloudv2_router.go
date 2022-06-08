package testserver

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

var ccloudv2Handlers = map[string]func(*testing.T) http.HandlerFunc{
	"/cmk/v2/clusters/{id}":            handleCmkCluster,
	"/cmk/v2/clusters":                 handleCmkClusters,
	"/iam/v2/api-keys/{id}":            handleIamApiKey,
	"/iam/v2/api-keys":                 handleIamApiKeys,
	"/iam/v2/users/{id}":               handleIamUser,
	"/iam/v2/users":                    handleIamUsers,
	"/iam/v2/service-accounts/{id}":    handleIamServiceAccount,
	"/iam/v2/service-accounts":         handleIamServiceAccounts,
	"/org/v2/environments/{id}":        handleOrgEnvironment,
	"/org/v2/environments":             handleOrgEnvironments,
	"/service-quota/v2/applied-quotas": handleAppliedQuotas,
}

type V2Router struct {
	*mux.Router
}

func NewV2Router(t *testing.T) *V2Router {
	router := &V2Router{
		Router: mux.NewRouter(),
	}
	router.buildV2Handler(t)
	return router
}

func (c *V2Router) buildV2Handler(t *testing.T) {
	for route, handler := range ccloudv2Handlers {
		c.HandleFunc(route, handler(t))
	}
}
