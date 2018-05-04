package http

import (
	"net/http"
	"golang.org/x/oauth2"
	"context"
	"fmt"
	"time"

	"github.com/confluentinc/cli/log"
)

const (
	timeout = time.Second * 10
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	Auth       *AuthService
	Connect    *ConnectService
	logger     *log.Logger
}

func NewClient(httpClient *http.Client, baseURL string, logger *log.Logger) *Client {
	client := &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
		logger:     logger,
	}
	client.Auth = NewAuthService(client)
	client.Connect = NewConnectService(client)
	return client
}

func NewClientWithJWT(ctx context.Context, jwt, baseURL string, logger *log.Logger) *Client {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, &http.Client{Timeout: timeout})
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: jwt})
	tc := oauth2.NewClient(ctx, ts)
	return NewClient(tc, baseURL, logger)
}

type confluentError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ConfluentError struct {
	OrgId   int            `json:"organization_id"`
	UserId  int            `json:"user_id"`
	Err     confluentError `json:"error"`
}

func (e *ConfluentError) Error() string {
	return fmt.Sprintf("confluent (%v): %v (org:%v, user:%v)", e.Err.Code, e.Err.Message, e.OrgId, e.UserId)
}
