package utils

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/confluentinc/cli/internal/pkg/log"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func SelfSignedCertClientFromPath(caCertPath string, logger *log.Logger) (*http.Client, error) {
	return CustomCAAndClientCertClient(caCertPath, "", "", logger)
}

func CustomCAAndClientCertClient(caCertPath string, clientCertPath string, clientKeyPath string, logger *log.Logger) (*http.Client, error) {
	var caCertReader *os.File
	if caCertPath != "" {
		caCertPath, err := filepath.Abs(caCertPath)
		if err != nil {
			return nil, err
		}
		logger.Debugf("Attempting to load certificate from absolute path %s", caCertPath)
		caCertReader, err = os.Open(caCertPath)
		if err != nil {
			return nil, err
		}
		defer caCertReader.Close()
		logger.Tracef("Successfully read CA certificate.")
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
		logger.Debugf("Attempting to load client key pair from absolute client cert path %s and absolute client key path %s", clientCertPath, clientKeyPath)
		cert, err = tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
		if err != nil {
			return nil, err
		}
	}

	logger.Tracef("Attempting to initialize HTTP client using certificates")
	client, err := SelfSignedCertClient(caCertReader, cert, logger)
	if err != nil {
		return nil, err
	}
	if caCertPath != "" {
		logger.Tracef("Successfully loaded certificate from %s", caCertPath)
	}
	if clientCertPath != "" {
		logger.Tracef("Successfully loaded certificate from %s", clientCertPath)
	}

	return client, nil
}

func SelfSignedCertClient(caCertReader io.Reader, clientCert tls.Certificate, logger *log.Logger) (*http.Client, error) {
	if caCertReader == nil && isEmptyClientCert(clientCert) {
		return nil, errors.New(errors.NoReaderForCustomCertErrorMsg)
	}

	var caCertPool *x509.CertPool
	if caCertReader != nil && caCertReader != (*os.File)(nil) {
		var err error
		caCertPool, err = x509.SystemCertPool() // load system certs
		if err != nil {
			logger.Warnf("Unable to load system certificates. Continuing with custom certificates only.")
		}
		logger.Tracef("Loaded certificate pool from system")
		if caCertPool == nil {
			logger.Tracef("(System certificate pool was blank)")
			caCertPool = x509.NewCertPool()
		}
		// read custom certs
		caCerts, err := ioutil.ReadAll(caCertReader)
		if err != nil {
			return nil, errors.Wrap(err, errors.ReadCertErrorMsg)
		}
		logger.Tracef("Specified ca certificate has been read")

		// Append custom certs to the system pool
		if ok := caCertPool.AppendCertsFromPEM(caCerts); !ok {
			return nil, errors.New(errors.NoCertsAppendedErrorMsg)
		}
		logger.Tracef("Successfully appended new certificate to the pool")
	}

	// Trust the updated cert pool in our client
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: caCertPool}
	logger.Tracef("Successfully created TLS config using certificate pool")

	if !isEmptyClientCert(clientCert) {
		transport.TLSClientConfig.Certificates = []tls.Certificate{clientCert}
		logger.Tracef("Successfully added client cert to TLS config")
	}

	defaultClient := DefaultClient()
	client := &http.Client{
		Transport:     transport,
		CheckRedirect: defaultClient.CheckRedirect,
		Jar:           defaultClient.Jar,
		Timeout:       defaultClient.Timeout,
	}
	logger.Tracef("Successfully set client properties")

	return client, nil
}

func isEmptyClientCert(cert tls.Certificate) bool {
	return cert.Certificate == nil && cert.Leaf == nil && cert.OCSPStaple == nil && cert.PrivateKey == nil && cert.SignedCertificateTimestamps == nil && cert.SupportedSignatureAlgorithms == nil
}

func DefaultClient() *http.Client {
	return http.DefaultClient
}
