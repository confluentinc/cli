//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/auth_mds_client.go --pkg mock --selfpkg github.com/confluentinc/cli mds_client.go MDSClientManager
package auth

import (
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/mds-sdk-go"
	"os"
	"path/filepath"
)

// Made it an interface so that we can inject MDS client for testing through GetMDSClient
type MDSClientManager interface {
	GetMDSClient(ctx *v3.Context) (*mds.APIClient, error)
}


type MDSClientManagerImpl struct {}


func (m *MDSClientManagerImpl) GetMDSClient(ctx *v3.Context) (*mds.APIClient, error){
	var mdsClient *mds.APIClient
	mdsConfig := mds.NewConfiguration()
	caCertPath := ctx.Platform.CaCertPath
	if ctx == nil || caCertPath == "" {
		mdsClient = mds.NewAPIClient(mdsConfig)
		return mdsClient, nil
	}
	caCertFile, err := getCertReader(caCertPath)
	if err != nil {
		return nil, err
	}
	defer caCertFile.Close()
	mdsConfig.HTTPClient, err = SelfSignedCertClient(caCertFile, ctx.Logger)
	if err != nil {
		return nil, err
	}
	// TODO: make sure it is the same logger as the one for the command
	ctx.Logger.Debugf("Successfully loaded certificate from %s", caCertPath)
	return mds.NewAPIClient(mdsConfig), nil
}

func getCertReader(caCertPath string) (*os.File, error) {
	caCertPath, err := filepath.Abs(caCertPath)
	if err != nil {
		return nil, err
	}
	caCertFile, err := os.Open(caCertPath)
	if err != nil {
		return nil, err
	}
	return caCertFile, nil
}
