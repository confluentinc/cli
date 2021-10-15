//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --dst ../../../mock/login_organization_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_organization_manager.go LoginOrganizationManager
package auth

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

type LoginOrganizationManager interface {
	GetLoginOrganizationFromArgs(cmd *cobra.Command) func() (string, error)
	GetLoginOrganizationFromEnvVar(cmd *cobra.Command) func() (string, error)
	GetDefaultLoginOrganization() func() (string, error)
}

type LoginOrganizationManagerImpl struct {
	logger *log.Logger
}

func GetLoginOrganization(getOrgFuncs ...func() (string, error)) (string, error) {
	var orgResourceId string
	var err error
	for _, getFunc := range getOrgFuncs {
		orgResourceId, err = getFunc()
		if err == nil && orgResourceId != "" {
			return orgResourceId, nil
		}
	}
	if err != nil {
		return "", err
	}
	return orgResourceId, nil
}

func NewLoginOrganizationManagerImp(logger *log.Logger) *LoginOrganizationManagerImpl {
	return &LoginOrganizationManagerImpl{ logger: logger }
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromArgs(cmd *cobra.Command) func() (string, error) {
	return func() (string, error) {
		return cmd.Flags().GetString("organization-id")
	}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromEnvVar(cmd *cobra.Command) func() (string, error) {
	return func() (string, error) {
		orgResourceId := os.Getenv(ConfluentCloudOrganizationIdEnvVar)
		if h.logger.GetLevel() >= log.WARN && orgResourceId != "" {
			utils.ErrPrintf(cmd, errors.FoundOrganizationIdMsg, orgResourceId, ConfluentCloudOrganizationIdEnvVar)
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
