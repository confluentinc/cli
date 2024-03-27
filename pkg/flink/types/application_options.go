package types

import (
	"github.com/confluentinc/cli/v3/pkg/config"
)

type ApplicationOptions struct {
	UnsafeTrace      bool   `json:"unsafeTrace"`
	UserAgent        string `json:"userAgent"`
	EnvironmentId    string `json:"environmentId"`
	EnvironmentName  string `json:"environmentName"`
	OrganizationId   string `json:"organizationId"`
	Database         string `json:"database"`
	ComputePoolId    string `json:"computePoolId"`
	ServiceAccountId string `json:"serviceAccountId"`
	Verbose          bool   `json:"verbose"`
	LSPBaseUrl       string `json:"lspBaseUrl"`
	GatewayURL       string `json:"gatewayUrl"`
	Context          *config.Context
}

func (a *ApplicationOptions) GetUnsafeTrace() bool {
	if a != nil {
		return a.UnsafeTrace
	}
	return false
}

func (a *ApplicationOptions) GetUserAgent() string {
	if a != nil {
		return a.UserAgent
	}
	return ""
}

func (a *ApplicationOptions) GetEnvironmentId() string {
	if a != nil {
		return a.EnvironmentId
	}
	return ""
}

func (a *ApplicationOptions) GetEnvironmentName() string {
	if a != nil {
		return a.EnvironmentName
	}
	return ""
}

func (a *ApplicationOptions) GetOrganizationId() string {
	if a != nil {
		return a.OrganizationId
	}
	return ""
}

func (a *ApplicationOptions) GetDatabase() string {
	if a != nil {
		return a.Database
	}
	return ""
}

func (a *ApplicationOptions) GetComputePoolId() string {
	if a != nil {
		return a.ComputePoolId
	}
	return ""
}

func (a *ApplicationOptions) GetServiceAccountId() string {
	if a != nil {
		return a.ServiceAccountId
	}
	return ""
}

func (a *ApplicationOptions) GetVerbose() bool {
	if a != nil {
		return a.Verbose
	}
	return false
}

func (a *ApplicationOptions) GetContext() *config.Context {
	if a != nil {
		return a.Context
	}
	return nil
}

func (a *ApplicationOptions) GetLSPBaseUrl() string {
	if a != nil {
		return a.LSPBaseUrl
	}
	return ""
}
