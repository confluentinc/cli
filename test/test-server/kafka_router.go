package test_server

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// kafka urls
const (
	aclsCreate      = "/2.0/kafka/{id}/acls"
	aclsList        = "/2.0/kafka/{cluster}/acls:search"
	aclsDelete      = "/2.0/kafka/{cluster}/acls/delete"
	link            = "/2.0/kafka/{cluster}/links/{link}"
	links           = "/2.0/kafka/{cluster}/links"
	topicMirrorStop = "/2.0/kafka/{cluster}/topics/{topic}/mirror:stop"
)

type KafkaRouter struct {
	*mux.Router
}

func NewKafkaRouter(t *testing.T) *KafkaRouter {
	router := NewEmptyKafkaRouter()
	router.buildKafkaHandler(t)
	return router
}

func NewEmptyKafkaRouter() *KafkaRouter {
	return &KafkaRouter{
		mux.NewRouter(),
	}
}

func (c *KafkaRouter) buildKafkaHandler(t *testing.T) {
	c.HandleFunc(aclsCreate, c.HandleKafkaACLsCreate(t))
	c.HandleFunc(aclsList, c.HandleKafkaACLsList(t))
	c.HandleFunc(aclsDelete, c.HandleKafkaACLsDelete(t))
	c.HandleFunc(link, c.HandleKafkaLink(t))
	c.HandleFunc(links, c.HandleKafkaLinks(t))
	c.HandleFunc(topicMirrorStop, c.HandleKafkaTopicMirrorStop(t))
	c.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}
