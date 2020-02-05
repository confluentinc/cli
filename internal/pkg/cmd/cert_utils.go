package cmd

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/confluentinc/cli/internal/pkg/log"
)

func SelfSignedCertClient(certReader io.Reader, logger *log.Logger) (*http.Client, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Warnf("Unable to load system certificates. Continuing with custom certificates only.")
	}
	if certPool == nil {
		certPool = x509.NewCertPool()
	}

	if certReader == nil {
		return nil, fmt.Errorf("no reader specified for reading custom certificates")
	}
	certs, err := ioutil.ReadAll(certReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %v", err)
	}

	// Append new cert to the system pool
	if ok := certPool.AppendCertsFromPEM(certs); !ok {
		return nil, fmt.Errorf("no certs appended, using system certs only")
	}

	// Trust the updated cert pool in our client
	transport := defaultTransport()
	transport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	client := DefaultClient()
	client.Transport = transport

	return client, nil
}

func defaultTransport() *http.Transport {
	// copied from the current net/http/transport.go dependency version, but it's already
	// out of date with respect to newer transport versions. For future proofing, this
	// should be replaced with:
	//     return http.defaultTransport.(*http.Transport).Clone()
	// but only after upgrading to go 1.13, since Clone isn't available until then.
	return &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
}

func DefaultClient() *http.Client {
	return http.DefaultClient
}
