//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/login_organization_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_organization_manager.go LoginOrganizationManager
package auth

import (
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type LoginOrganizationManager interface {
	GetLoginOrganizationFromArgs(*cobra.Command) func() (string, error)
	GetLoginOrganizationFromEnvVar(*cobra.Command) func() (string, error)
	GetDefaultLoginOrganization() func() (string, error)
}

type LoginOrganizationManagerImpl struct{}

func GetLoginOrganization(getOrgFuncs ...func() (string, error)) (string, error) {
	var multiErr error
	for _, getFunc := range getOrgFuncs {
		orgResourceId, err := getFunc()
		if err == nil && orgResourceId != "" {
			return orgResourceId, nil
		} else if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}
	return "", multiErr
}

func NewLoginOrganizationManagerImpl() *LoginOrganizationManagerImpl {
	return &LoginOrganizationManagerImpl{}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromArgs(cmd *cobra.Command) func() (string, error) {
	return func() (string, error) {
		return cmd.Flags().GetString("organization-id")
	}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromEnvVar(cmd *cobra.Command) func() (string, error) {
	return func() (string, error) {
		orgResourceId := os.Getenv(ConfluentCloudOrganizationId)
		if orgResourceId != "" {
			log.CliLogger.Warnf(errors.FoundOrganizationIdMsg, orgResourceId, ConfluentCloudOrganizationId)
		}
		return orgResourceId, nil
	}
}

func (h *LoginOrganizationManagerImpl) GetDefaultLoginOrganization() func() (string, error) {
	return func() (string, error) {
		// empty org resource id will be interpreted as the default org by the login API
		return "", nil
	}
}
