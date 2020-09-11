package auth

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func SelfSignedCertClient(certReader io.Reader, logger *log.Logger) (*http.Client, error) {
	certPool, err := x509.SystemCertPool()
	if err != nil {
		logger.Warnf("Unable to load system certificates. Continuing with custom certificates only.")
	}
	logger.Tracef("Loaded certificate pool from system")
	if certPool == nil {
		logger.Tracef("(System certificate pool was blank)")
		certPool = x509.NewCertPool()
	}

	if certReader == nil {
		return nil, errors.New(errors.NoReaderForCustomCertErrorMsg)
	}
	certs, err := ioutil.ReadAll(certReader)
	if err != nil {
		return nil, errors.Wrap(err, errors.ReadCertErrorMsg)
	}
	logger.Tracef("Specified certificate has been read")

	// Append new cert to the system pool
	if ok := certPool.AppendCertsFromPEM(certs); !ok {
		return nil, errors.New(errors.NoCertsAppendedErrorMsg)
	}

	logger.Tracef("Successfully appended new certificate to the pool")

	// Trust the updated cert pool in our client
	transport := defaultTransport()
	transport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	logger.Tracef("Successfully created TLS config using certificate pool")
	client := DefaultClient()
	client.Transport = transport
	logger.Tracef("Successfully set client properties")

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
