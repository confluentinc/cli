//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_mds_client.go --pkg mock --selfpkg github.com/confluentinc/cli/v3 mds_client.go MDSClientManager
package auth

import (
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

// Made it an interface so that we can inject MDS client for testing through GetMDSClient
type MDSClientManager interface {
	GetMDSClient(url, caCertPath string, unsafeTrace bool) (*mdsv1.APIClient, error)
}

type MDSClientManagerImpl struct{}

func (m *MDSClientManagerImpl) GetMDSClient(url, caCertPath string, unsafeTrace bool) (*mdsv1.APIClient, error) {
	mdsConfig := mdsv1.NewConfiguration()
	mdsConfig.Debug = unsafeTrace

	if caCertPath != "" {
		log.CliLogger.Debugf("CA certificate path was specified.  Note, the set of supported ciphers for the CLI can be found at https://golang.org/pkg/crypto/tls/#pkg-constants")
		var err error

		mdsConfig.HTTPClient, err = utils.SelfSignedCertClientFromPath(caCertPath)
		if err != nil {
			return nil, err
		}
	} else {
		mdsConfig.HTTPClient = utils.DefaultClient()
	}
	mdsClient := mdsv1.NewAPIClient(mdsConfig)
	mdsClient.ChangeBasePath(url)
	return mdsClient, nil
}
