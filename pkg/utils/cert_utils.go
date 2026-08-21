package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func GetCAAndClientCertClient(caCertPath, clientCertPath, clientKeyPath string) (*http.Client, error) {
	caCertPath, err := filepath.Abs(caCertPath)
	if err != nil {
		return nil, err
	}
	log.CliLogger.Debugf("Attempting to load certificate from absolute path %s", caCertPath)
	caCertReader, err := os.Open(caCertPath)
	if err != nil {
		return nil, errors.NewErrorWithSuggestions(
			"no Certificate Authority certificate specified",
			"Please specify `--certificate-authority-path` to enable Schema Registry client.",
		)
	}
	defer caCertReader.Close()
	log.CliLogger.Tracef("Successfully read CA certificate")

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

	client, err := SelfSignedCertClient(caCertReader, cert)
	if err != nil {
		return nil, err
	}
	return client, nil
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
		return nil, fmt.Errorf("no reader specified for reading custom certificates")
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
			return nil, fmt.Errorf("failed to read certificate: %w", err)
		}
		log.CliLogger.Tracef("Specified CA certificate has been read")

		// Append custom certs to the system pool
		if ok := caCertPool.AppendCertsFromPEM(caCerts); !ok {
			return nil, fmt.Errorf("no certs appended, using system certs only")
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

// Could refactor the above CustomCAAndClientCertClient to use this, but for now leaving it separate to avoid breaking changes
func GetEnrichedCACertPool(caCertPath string) (*x509.CertPool, error) {
	// Load system certs (or initialize a new one if unable to load system) as a certificate pool
	caCertPool, err := x509.SystemCertPool()
	if err != nil {
		log.CliLogger.Warnf("Unable to load system certificates; continuing with custom certificates only")
	}
	log.CliLogger.Tracef("Loaded certificate pool from system")
	if caCertPool == nil {
		log.CliLogger.Tracef("(System certificate pool was blank)")
		caCertPool = x509.NewCertPool()
	}

	// If the provided path is not empty, and is a valid file, add it to the certificate pool
	if caCertPath == "" {
		log.CliLogger.Tracef("No custom CA certificate specified, using system certs only")
		return caCertPool, nil
	}

	// Validate and read the custom certificate file
	absPath, err := filepath.Abs(caCertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve certificate path: %w", err)
	}

	log.CliLogger.Debugf("Attempting to load certificate from absolute path %s", absPath)
	caCertFile, err := os.Open(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file: %w", err)
	}
	defer caCertFile.Close()

	customCaCerts, err := io.ReadAll(caCertFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate: %w", err)
	}
	log.CliLogger.Tracef("Successfully read CA certificate")

	// Append custom certs to the system pool
	if ok := caCertPool.AppendCertsFromPEM(customCaCerts); !ok {
		return nil, fmt.Errorf("no valid certificates found in file: %s", absPath)
	}
	log.CliLogger.Tracef("Successfully appended new certificate to the pool")

	return caCertPool, nil
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
