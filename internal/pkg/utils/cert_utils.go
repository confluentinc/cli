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
	caCertPath, err := filepath.Abs(caCertPath)
	if err != nil {
		return nil, err
	}

	logger.Debugf("Attempting to load certificate from absolute path %s", caCertPath)
	certReader, err := os.Open(caCertPath)
	if err != nil {
		return nil, err
	}
	defer certReader.Close()
	logger.Tracef("Successfully read CA certificate.")

	logger.Tracef("Attempting to initialize HTTP client using certificate")
	client, err := SelfSignedCertClient(certReader, logger)
	if err != nil {
		return nil, err
	}
	logger.Tracef("Successfully loaded certificate from %s", caCertPath)
	return client, nil
}

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
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{RootCAs: certPool}
	logger.Tracef("Successfully created TLS config using certificate pool")
	client := DefaultClient()
	client.Transport = transport
	logger.Tracef("Successfully set client properties")

	return client, nil
}

func DefaultClient() *http.Client {
	return http.DefaultClient
}
