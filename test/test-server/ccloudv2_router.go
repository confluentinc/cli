package test_server

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

var ccloudv2Handlers = map[string]func(*testing.T) http.HandlerFunc{
	"/cmk/v2/clusters/{id}":            HandleCmkCluster,
	"/cmk/v2/clusters":                 HandleCmkClusters,
	"/iam/v2/users/{id}":               HandleIamUser,
	"/iam/v2/users":                    HandleIamUsers,
	"/iam/v2/service-accounts/{id}":    HandleIamServiceAccount,
	"/iam/v2/service-accounts":         HandleIamServiceAccounts,
	"/org/v2/environments/{id}":        HandleOrgEnvironment,
	"/org/v2/environments":             HandleOrgEnvironments,
	"/service-quota/v2/applied-quotas": HandleAppliedQuotas,
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
