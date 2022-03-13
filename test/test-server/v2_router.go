package test_server

import (
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
	c.addIamClusterRoutes(t)
	c.addIamServiceAccountRoutes(t)
}

func (c *V2Router) addIamClusterRoutes(t *testing.T) {
	c.HandleFunc(iamUser, c.HandleIamUser(t))
	c.HandleFunc(iamUsers, c.HandleIamUsers(t))
}

func (c *V2Router) addIamServiceAccountRoutes(t *testing.T) {
	c.HandleFunc(iamServiceAccount, c.HandleIamServiceAccount(t))
	c.HandleFunc(iamServiceAccounts, c.HandleIamServiceAccounts(t))
}
