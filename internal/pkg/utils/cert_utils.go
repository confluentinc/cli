package utils

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

func GetCAClient(caCertPath string) (*http.Client, error) {
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		return nil, errors.NewErrorWithSuggestions(errors.CaCertNotSpecifiedErrorMsg, errors.SRCaCertSuggestions)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	transport := DefaultTransport()
	transport.TLSClientConfig.RootCAs = caCertPool
	return DefaultClientWithTransport(transport), nil
}

func SelfSignedCertClientFromPath(caCertPath string) (*http.Client, error) {
	return CustomCAAndClientCertClient(caCertPath, "", "")
}

func CustomCAAndClientCertClient(caCertPath, clientCertPath, clientKeyPath string) (*http.Client, error) {
	var caCertReader *os.File
	if caCertPath != "" {
		caCertPath, err := filepath.Abs(caCertPath)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Debugf("Attempting to load certificate from absolute path %s", caCertPath)
		caCertReader, err = os.Open(caCertPath)
		if err != nil {
			return nil, err
		}
		defer caCertReader.Close()
		log.CliLogger.Tracef("Successfully read CA certificate")
	}
	var cert tls.Certificate
	if clientCertPath != "" {
		clientCertPath, err := filepath.Abs(clientCertPath)
		if err != nil {
			return nil, err
		}
		clientKeyPath, err = filepath.Abs(clientKeyPath)
		if err != nil {
			return nil, err
		}
		log.CliLogger.Debugf("Attempting to load client key pair from absolute client cert path %s and absolute client key path %s", clientCertPath, clientKeyPath)
		cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
	}

	log.CliLogger.Tracef("Attempting to initialize HTTP client using certificates")
	client, err := SelfSignedCertClient(caCertReader, cert)
	if err != nil {
		return nil, err
	}
	if caCertPath != "" {
		log.CliLogger.Tracef("Successfully loaded certificate from %s", caCertPath)
	}
	if clientCertPath != "" {
		log.CliLogger.Tracef("Successfully loaded certificate from %s", clientCertPath)
	}

	return client, nil
}

func SelfSignedCertClient(caCertReader io.Reader, clientCert tls.Certificate) (*http.Client, error) {
	if caCertReader == nil && isEmptyClientCert(clientCert) {
		return nil, errors.New(errors.NoReaderForCustomCertErrorMsg)
	}
	transport := DefaultTransport()

	var caCertPool *x509.CertPool
	if caCertReader != nil && caCertReader != (*os.File)(nil) {
		var err error
		caCertPool, err = x509.SystemCertPool() // load system certs
		if err != nil {
			log.CliLogger.Warnf("Unable to load system certificates; continuing with custom certificates only")
		}
		log.CliLogger.Tracef("Loaded certificate pool from system")
		if caCertPool == nil {
			log.CliLogger.Tracef("(System certificate pool was blank)")
			caCertPool = x509.NewCertPool()
		}
		// read custom certs
		caCerts, err := io.ReadAll(caCertReader)
		if err != nil {
			return nil, errors.Wrap(err, errors.ReadCertErrorMsg)
		}
		log.CliLogger.Tracef("Specified CA certificate has been read")

		// Append custom certs to the system pool
		if ok := caCertPool.AppendCertsFromPEM(caCerts); !ok {
			return nil, errors.New(errors.NoCertsAppendedErrorMsg)
		}
		log.CliLogger.Tracef("Successfully appended new certificate to the pool")
		// Trust the updated cert pool in our client
		transport.TLSClientConfig.RootCAs = caCertPool
		log.CliLogger.Tracef("Successfully created TLS config using certificate pool")
	}

	if !isEmptyClientCert(clientCert) {
		transport.TLSClientConfig.Certificates = []tls.Certificate{clientCert}
		log.CliLogger.Tracef("Successfully added client certificate to TLS config")
	}
	defaultClient := DefaultClient()
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: defaultClient.CheckRedirect,
		Jar:           defaultClient.Jar,
		Timeout:       defaultClient.Timeout,
	}

	log.CliLogger.Tracef("Successfully set client properties")

	return client, nil
}

func isEmptyClientCert(cert tls.Certificate) bool {
	return cert.Certificate == nil && cert.Leaf == nil && cert.OCSPStaple == nil && cert.PrivateKey == nil && cert.SignedCertificateTimestamps == nil && cert.SupportedSignatureAlgorithms == nil
}

func DefaultTransport() *http.Transport {
	return &http.Transport{
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
}

func DefaultClient() *http.Client {
	return &http.Client{
		Transport: DefaultTransport(),
	}
}

func DefaultClientWithTransport(transport *http.Transport) *http.Client {
	return &http.Client{Transport: transport}
}
