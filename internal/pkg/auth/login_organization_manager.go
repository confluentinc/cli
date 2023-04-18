//go:generate mocker --dst ../../../mock/login_organization_manager.go --pkg mock --selfpkg github.com/confluentinc/cli login_organization_manager.go LoginOrganizationManager --prefix ""
package auth

import (
	"os"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/log"
)

type LoginOrganizationManager interface {
	GetLoginOrganizationFromFlag(*cobra.Command) func() string
	GetLoginOrganizationFromEnvironmentVariable() func() string
	GetLoginOrganizationFromConfigurationFile(cfg *v1.Config) func() string
}

type LoginOrganizationManagerImpl struct{}

func GetLoginOrganization(getOrgFuncs ...func() string) string {
	for _, getFunc := range getOrgFuncs {
		if id := getFunc(); id != "" {
			return id
		}
	}

	// An empty organization ID is interpreted as the default organization by the login API
	return ""
}

func NewLoginOrganizationManagerImpl() *LoginOrganizationManagerImpl {
	return &LoginOrganizationManagerImpl{}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromFlag(cmd *cobra.Command) func() string {
	return func() string {
		organizationId, _ := cmd.Flags().GetString("organization-id")
		return organizationId
	}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromEnvironmentVariable() func() string {
	return func() string {
		organizationId := os.Getenv(ConfluentCloudOrganizationId)
		if organizationId != "" {
			log.CliLogger.Debugf(`Found default organization ID "%s" from environment variable "%s"`, organizationId, ConfluentCloudOrganizationId)
		}
		return organizationId
	}
}

func (h *LoginOrganizationManagerImpl) GetLoginOrganizationFromConfigurationFile(cfg *v1.Config) func() string {
	return func() string {
		return cfg.Context().GetCurrentOrganization()
	}
}
