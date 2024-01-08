package ccloud

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/dghubble/sling"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	timeout                = time.Second * 10
	ccloudSDKGoPackageName = "ccloud-sdk-go-v1-public"
)

var (
	baseTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		ForceAttemptHTTP2: true,
	}

	// BaseClient represents a raw golang http client with the SDK http defaults.
	BaseClient = &http.Client{
		Timeout:   timeout,
		Transport: baseTransport,
	}

	// Requests should be serialized as JSONPB by default, specify something else if you want otherwise
	Request = NewJSONPBBodyProvider
)

type Params struct {
	BaseURL        string
	UserAgent      string
	BaseClient     *http.Client // all sling clients should extend from this (e.g., kafka-api with diff auth token)
	HttpClient     *http.Client // this is the http client for the cloud APIs
	Logger         Logger
	MetricsBaseURL string
}

// Client represents the Confluent SDK client.
type Client struct {
	*Params
	sling               *sling.Sling
	Account             AccountInterface
	Auth                Auth
	Billing             Billing
	EnvironmentMetadata EnvironmentMetadata
	ExternalIdentity    ExternalIdentity
	Growth              Growth
	SchemaRegistry      SchemaRegistry
	Signup              Signup
	User                UserInterface
}

func GetSlingWithNewClient(s *sling.Sling, client *http.Client, logger Logger) *sling.Sling {
	return s.New().Doer(newLoggingHttpClient(client, logger))
}

// Proxy for HTTP client for logging
type LoggingHttpClient struct {
	client *http.Client
	logger Logger
}

func newLoggingHttpClient(client *http.Client, logger Logger) *LoggingHttpClient {
	if client == nil {
		client = BaseClient
	}
	return &LoggingHttpClient{client: client, logger: logger}
}

func (c *LoggingHttpClient) Do(req *http.Request) (*http.Response, error) {
	caller, err := getSDKMethodCaller()
	if err != nil {
		c.logger.Debugf("unable to obtain ccloud-sdk-go-v1-public service caller method: %s", err.Error())
	}
	type readCloser struct {
		io.Reader
		io.Closer
	}
	reqLog := fmt.Sprintf("request: %s %s", req.Method, req.URL)
	if req.Body != nil {
		body := fmt.Sprintf("%s", req.Body)
		reqLog = fmt.Sprintf("%s Body:%s", reqLog, body[1:len(body)-1])
	}
	c.logger.Debug(caller, " ", reqLog)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	bufferReader := bytes.NewBuffer(b)
	res.Body = readCloser{bufferReader, res.Body}
	resLog := fmt.Sprintf("response: %s X-Request-Id:%s Body: %s", res.Status, res.Header.Get("X-Request-Id"), bufferReader)
	c.logger.Debug(caller, " ", resLog, " ", reqLog)
	return res, nil
}

// NewClient creates a Confluent SDK client.
func NewClient(p *Params) *Client {
	p = setDefaults(p)
	client := &Client{
		Params: p,
		sling: sling.New().
			Doer(newLoggingHttpClient(p.HttpClient, p.Logger)).
			Base(p.BaseURL).
			Set("User-Agent", p.UserAgent).
			ResponseDecoder(NewJSONPBDecoder()),
	}
	client.Auth = NewAuthService(client)
	client.Account = NewAccountService(client)
	client.Billing = NewBillingService(client)
	client.EnvironmentMetadata = NewEnvironmentMetadataService(client)
	client.ExternalIdentity = NewExternalIdentityService(client)
	client.Growth = NewGrowthService(client)
	client.SchemaRegistry = NewSchemaRegistryService(client)
	client.Signup = NewSignupService(client)
	client.User = NewUserService(client)
	return client
}

// NewClientWithJWT creates a Confluent SDK client which authenticates with the given JSON Web Token (JWT).
func NewClientWithJWT(ctx context.Context, jwt string, p *Params) *Client {
	p = setDefaults(p)
	p.HttpClient = InjectOAuth(ctx, jwt, p.BaseClient)
	return NewClient(p)
}

func InjectOAuth(ctx context.Context, accessToken string, baseClient *http.Client) *http.Client {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, baseClient)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tc := oauth2.NewClient(ctx, ts)
	return tc
}

func setDefaults(p *Params) *Params {
	if p == nil {
		p = &Params{}
	}
	if p.BaseURL == "" {
		p.BaseURL = "https://confluent.cloud"
	}
	if p.Logger == nil {
		p.Logger = logrus.New()
	}
	if p.BaseClient == nil {
		p.BaseClient = BaseClient
	}
	if p.UserAgent != "" {
		p.UserAgent += " "
	}
	p.UserAgent += fmt.Sprintf("ccloud-sdk-go-v1-public/%s (%s/%s; %s)", SDKVersion, runtime.GOOS, runtime.GOARCH, runtime.Version())
	if p.MetricsBaseURL == "" {
		p.MetricsBaseURL = "https://api.telemetry.confluent.cloud"
	}
	return p
}

func getSDKMethodCaller() (string, error) {
	functionName, err := getCCloudSDKGoFunctionName()
	if err != nil {
		return "", err
	}
	return extractCallerName(functionName), nil
}

func getCCloudSDKGoFunctionName() (string, error) {
	ptr, index, err := getCCloudSDKGoPackageFrame()
	if err != nil {
		return "", err
	}
	functionName := getFunctionNameFromFramePtr(ptr)

	// workaround for when sling call is in helper function
	if !strings.Contains(functionName, ccloudSDKGoPackageName) {
		functionName, err = getCCloudSDKGoFunctionNameInNextFrame(index)
		if err != nil {
			return "", err
		}
	}

	return functionName, nil
}

func getFunctionNameFromFramePtr(ptr uintptr) string {
	ptrSlice := []uintptr{ptr}
	frame, _ := runtime.CallersFrames(ptrSlice).Next()
	return frame.Function
}

func getCCloudSDKGoPackageFrame() (uintptr, int, error) {
	// Index start at 4 to skip the frames for getSDKMethodCaller, getFunctionName, getCCLOUDSDKGOPacakgeFramePC,
	// and the Do method of the LoggingHttpClient which are all in ccloud-sdk-go-v1-public
	index := 4
	ptr, fileName, _, ok := runtime.Caller(index)
	for ok {
		match := strings.Contains(fileName, ccloudSDKGoPackageName)
		if match {
			break
		}
		index += 1
		ptr, fileName, _, ok = runtime.Caller(index)
	}
	if !ok {
		return 0, 0, fmt.Errorf("frame index out of bound")
	}
	return ptr, index, nil
}

// workaround for when sling network calls are made in a helper function instead of directly in a ccloud-sdk-go-v1-public service method
// e.g. kafka calls that make sling call in setKafkaAPI helper method
func getCCloudSDKGoFunctionNameInNextFrame(index int) (string, error) {
	ptr, _, _, ok := runtime.Caller(index + 1)
	if !ok {
		return "", fmt.Errorf("frame index out of bound")
	}
	functionName := getFunctionNameFromFramePtr(ptr)
	if !strings.Contains(functionName, ccloudSDKGoPackageName) {
		return "", fmt.Errorf("could not find ccloud-sdk-go-v1-public service function name")
	}
	return functionName, nil
}

// Extract caller name from <path>.(*<Service Name>).<Method Name> to <Service Name>.<Method Name>
func extractCallerName(fname string) string {
	fnameSplit := strings.Split(fname, "/")
	fnameSplit = strings.Split(fnameSplit[len(fnameSplit)-1], ".")
	serviceName := fnameSplit[1]
	serviceName = serviceName[2 : len(serviceName)-1]
	return serviceName + "." + fnameSplit[2]
}
