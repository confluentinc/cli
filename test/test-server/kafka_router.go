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
	aclsCreate      = "/2.0/kafka/{cluster}/acls"
	aclsList        = "/2.0/kafka/{cluster}/acls:search"
	aclsDelete      = "/2.0/kafka/{cluster}/acls/delete"
	link            = "/2.0/kafka/{cluster}/links/{link}"
	links           = "/2.0/kafka/{cluster}/links"
	topicMirrorStop = "/2.0/kafka/{cluster}/topics/{topic}/mirror:stop"
	topics          = "/2.0/kafka/{cluster}/topics"
	topic           = "/2.0/kafka/{cluster}/topics/{topic}"
	topicConfig     = "/2.0/kafka/{cluster}/topics/{topic}/config"
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

func (k *KafkaRouter) buildKafkaHandler(t *testing.T) {
	k.HandleFunc(aclsCreate, k.HandleKafkaACLsCreate(t))
	k.HandleFunc(aclsList, k.HandleKafkaACLsList(t))
	k.HandleFunc(aclsDelete, k.HandleKafkaACLsDelete(t))
	k.HandleFunc(link, k.HandleKafkaLink(t))
	k.HandleFunc(links, k.HandleKafkaLinks(t))
	k.HandleFunc(topicMirrorStop, k.HandleKafkaTopicMirrorStop(t))
	k.HandleFunc(topics, k.HandleKafkaListCreateTopic(t))
	k.HandleFunc(topic, k.HandleKafkaDescribeDeleteTopic(t))
	k.HandleFunc(topicConfig, k.HandleKafkaTopicListConfig(t))
	k.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}
