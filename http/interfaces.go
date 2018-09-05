// go:generate mocker --prefix "" --out mock/http_auth.go --pkg mock http Auth
// go:generate mocker --prefix "" --out mock/http_user.go --pkg mock http User
// go:generate mocker --prefix "" --out mock/http_apikey.go --pkg mock http APIKey
// go:generate mocker --prefix "" --out mock/http_kafka.go --pkg mock http Kafka
// go:generate mocker --prefix "" --out mock/http_connect.go --pkg mock http Connect

package http

import (
	"net/http"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/confluentinc/cli/shared"
)

type Auth interface {
	Login(username, password string) (string, error)
	User() (*shared.AuthConfig, error)
}

type User interface {
	List() ([]*orgv1.User, *http.Response, error)
	Describe(user *orgv1.User) (*orgv1.User, *http.Response, error)
}

type APIKey interface {
	Create(key *orgv1.ApiKey) (*orgv1.ApiKey, *http.Response, error)
}

type Kafka interface {
	List(cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, *http.Response, error)
	Describe(cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, *http.Response, error)
}

type Connect interface {
	List(cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, *http.Response, error)
	Describe(cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, *http.Response, error)
	DescribeS3Sink(cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	CreateS3Sink(config *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	UpdateS3Sink(cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	Delete(cluster *schedv1.ConnectCluster) (*http.Response, error)
}

