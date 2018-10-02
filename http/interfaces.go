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

// Auth allows authenticating in Confluent Cloud
type Auth interface {
	Login(username, password string) (string, error)
	User() (*shared.AuthConfig, error)
}

// User service allows managing users in Confluent Cloud
type User interface {
	List() ([]*orgv1.User, *http.Response, error)
	Describe(user *orgv1.User) (*orgv1.User, *http.Response, error)
}

// APIKey service allows managing API Keys in Confluent Cloud
type APIKey interface {
	Create(key *schedv1.ApiKey) (*schedv1.ApiKey, *http.Response, error)
}

// Kafka service allows managing Kafka clusters in Confluent Cloud
type Kafka interface {
	List(cluster *schedv1.KafkaCluster) ([]*schedv1.KafkaCluster, *http.Response, error)
	Describe(cluster *schedv1.KafkaCluster) (*schedv1.KafkaCluster, *http.Response, error)
	Create(config *schedv1.KafkaClusterConfig) (*schedv1.KafkaCluster, *http.Response, error)
	Delete(cluster *schedv1.KafkaCluster) (*http.Response, error)
}

// Connect service allows managing Connect clusters in Confluent Cloud
type Connect interface {
	List(cluster *schedv1.ConnectCluster) ([]*schedv1.ConnectCluster, *http.Response, error)
	Describe(cluster *schedv1.ConnectCluster) (*schedv1.ConnectCluster, *http.Response, error)
	DescribeS3Sink(cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	CreateS3Sink(config *schedv1.ConnectS3SinkClusterConfig) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	UpdateS3Sink(cluster *schedv1.ConnectS3SinkCluster) (*schedv1.ConnectS3SinkCluster, *http.Response, error)
	Delete(cluster *schedv1.ConnectCluster) (*http.Response, error)
}

// KSQL service allows managing KSQL clusters in Confluent Cloud
type KSQL interface {
	List(cluster *schedv1.KSQLCluster) ([]*schedv1.KSQLCluster, *http.Response, error)
	Describe(cluster *schedv1.KSQLCluster) (*schedv1.KSQLCluster, *http.Response, error)
	Create(config *schedv1.KSQLClusterConfig) (*schedv1.KSQLCluster, *http.Response, error)
	Delete(cluster *schedv1.KSQLCluster) (*http.Response, error)
}
