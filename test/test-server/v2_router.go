package test_server

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

// v2 urls
const (
	iamUser            = "/iam/v2/users/{id}"
	iamUsers           = "/iam/v2/users"
	iamServiceAccount  = "/iam/v2/service-accounts/{id}"
	iamServiceAccounts = "/iam/v2/service-accounts"
)

var iamHandlers = map[string]func(*V2Router, *testing.T) func(http.ResponseWriter, *http.Request){
	iamUser:            (*V2Router).HandleIamUser,
	iamUsers:           (*V2Router).HandleIamUsers,
	iamServiceAccount:  (*V2Router).HandleIamServiceAccount,
	iamServiceAccounts: (*V2Router).HandleIamServiceAccounts,
}

type V2Router struct {
	*mux.Router
}

func NewV2Router(t *testing.T) *V2Router {
	router := NewEmptyV2Router()
	router.buildV2Handler(t)
	return router
}

func NewEmptyV2Router() *V2Router {
	return &V2Router{
		Router: mux.NewRouter(),
	}
}

func (c *V2Router) buildV2Handler(t *testing.T) {
	for route, handler := range iamHandlers {
		c.HandleFunc(route, handler(c, t))
	}
}
