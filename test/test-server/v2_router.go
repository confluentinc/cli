package test_server

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

var iamHandlers = map[string]func(*testing.T) http.HandlerFunc{
	"/iam/v2/users/{id}":            HandleIamUser,
	"/iam/v2/users":                 HandleIamUsers,
	"/iam/v2/service-accounts/{id}": HandleIamServiceAccount,
	"/iam/v2/service-accounts":      HandleIamServiceAccounts,
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
	for route, handler := range iamHandlers {
		c.HandleFunc(route, handler(t))
	}
}
