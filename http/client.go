package http

import (
	"context"
	"net/http"
	"time"

	"github.com/dghubble/sling"
	"golang.org/x/oauth2"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	"github.com/confluentinc/cli/log"
)

const (
	timeout = time.Second * 10
)

var (
	BaseClient  = &http.Client{Timeout: timeout}
	errNotFound = &corev1.Error{Code: http.StatusNotFound, Message: "cluster not found"} // matches gateway response
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	logger     *log.Logger
	sling      *sling.Sling
	Auth       *AuthService
	Kafka      *KafkaService
	Connect    *ConnectService
}

func NewClient(httpClient *http.Client, baseURL string, logger *log.Logger) *Client {
	client := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     logger,
		sling:      sling.New().Client(httpClient).Base(baseURL).Decoder(NewJSONPBDecoder()),
	}
	client.Auth = NewAuthService(client)
	client.Kafka = NewKafkaService(client)
	client.Connect = NewConnectService(client)
	return client
}

func NewClientWithJWT(ctx context.Context, jwt, baseURL string, logger *log.Logger) *Client {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, BaseClient)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwt})
	tc := oauth2.NewClient(ctx, ts)
	return NewClient(tc, baseURL, logger)
}
