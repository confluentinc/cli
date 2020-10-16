//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_mds_client.go --pkg mock --selfpkg github.com/confluentinc/cli mds_client.go MDSClientManager
package auth

import (
	"os"
	"path/filepath"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	log "github.com/confluentinc/cli/internal/pkg/log"
)

// Made it an interface so that we can inject MDS client for testing through GetMDSClient
type MDSClientManager interface {
	GetMDSClient(ctx *v3.Context, caCertPath string, caCertPathFlagChanged bool, url string, logger *log.Logger) (*mds.APIClient, error)
}

type MDSClientManagerImpl struct{}

func (m *MDSClientManagerImpl) GetMDSClient(ctx *v3.Context, caCertPath string, caCertPathFlagChanged bool, url string, logger *log.Logger) (*mds.APIClient, error) {
	mdsClient := initializeMDS(ctx, logger)
	if logger.GetLevel() == log.DEBUG || logger.GetLevel() == log.TRACE {
		mdsClient.GetConfig().Debug = true
	}
	if caCertPathFlagChanged {
		if caCertPath == "" {
			// revert to default client regardless of previously configured client
			mdsClient.GetConfig().HTTPClient = DefaultClient()
		} else {
			logger.Debugf("CA certificate path was specified.  Note, the set of supported ciphers for the CLI can be found at https://golang.org/pkg/crypto/tls/#pkg-constants")
			// override previously configured httpclient if a new cert path was specified
			certReader, err := getCertReader(caCertPath, logger)
			if err != nil {
				return nil, err
			}
			logger.Tracef("Successfully read CA certificate.")
			logger.Tracef("Attempting to initialize HTTP client using certificate")
			mdsClient.GetConfig().HTTPClient, err = SelfSignedCertClient(certReader, logger)
			if err != nil {
				return nil, err
			}
			logger.Tracef("Successfully loaded certificate from %s", caCertPath)
		}
	}
	mdsClient.ChangeBasePath(url)
	return mdsClient, nil
}

func initializeMDS(ctx *v3.Context, logger *log.Logger) *mds.APIClient {
	mdsConfig := mds.NewConfiguration()
	if ctx == nil || ctx.Platform.CaCertPath == "" {
		return mds.NewAPIClient(mdsConfig)
	}
	caCertPath := ctx.Platform.CaCertPath
	// Try to load certs. On failure, warn, but don't error out because this may be an auth command, so there may
	// be a --ca-cert-path flag on the cmd line that'll fix whatever issue there is with the cert file in the config
	caCertFile, err := os.Open(caCertPath)
	if err == nil {
		defer caCertFile.Close()
		mdsConfig.HTTPClient, err = SelfSignedCertClient(caCertFile, logger)
	}
	if err != nil {
		logger.Warnf("Unable to load certificate from %s. %s. Resulting SSL errors will be fixed by logging in with the --ca-cert-path flag.", caCertPath, err.Error())
		mdsConfig.HTTPClient = DefaultClient()
	}
	return mds.NewAPIClient(mdsConfig)
}

func getCertReader(caCertPath string, logger *log.Logger) (*os.File, error) {
	caCertPath, err := filepath.Abs(caCertPath)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Attempting to load certificate from absolute path %s", caCertPath)
	caCertFile, err := os.Open(caCertPath)
	if err != nil {
		return nil, err
	}
	return caCertFile, nil
}
