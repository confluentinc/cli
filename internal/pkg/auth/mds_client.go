//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_mds_client.go --pkg mock --selfpkg github.com/confluentinc/cli mds_client.go MDSClientManager
package auth

import (
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	log "github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

// Made it an interface so that we can inject MDS client for testing through GetMDSClient
type MDSClientManager interface {
	GetMDSClient(url string, caCertPath string, logger *log.Logger) (*mds.APIClient, error)
}

type MDSClientManagerImpl struct{}

func (m *MDSClientManagerImpl) GetMDSClient(url string, caCertPath string, logger *log.Logger) (*mds.APIClient, error) {
	mdsConfig := mds.NewConfiguration()
	if logger.GetLevel() == log.DEBUG || logger.GetLevel() == log.TRACE {
		mdsConfig.Debug = true
	}
	if caCertPath != "" {
		logger.Debugf("CA certificate path was specified.  Note, the set of supported ciphers for the CLI can be found at https://golang.org/pkg/crypto/tls/#pkg-constants")
		var err error

		mdsConfig.HTTPClient, err = utils.SelfSignedCertClientFromPath(caCertPath, logger)
		if err != nil {
			return nil, err
		}
	} else {
		mdsConfig.HTTPClient = utils.DefaultClient()
	}
	mdsClient := mds.NewAPIClient(mdsConfig)
	mdsClient.ChangeBasePath(url)
	return mdsClient, nil
}
