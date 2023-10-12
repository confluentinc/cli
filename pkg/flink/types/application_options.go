package types

import (
	"github.com/confluentinc/cli/v3/pkg/config"
)

type ApplicationOptions struct {
	UnsafeTrace      bool
	UserAgent        string
	EnvironmentId    string
	EnvironmentName  string
	OrganizationId   string
	Database         string
	ComputePoolId    string
	ServiceAccountId string
	Verbose          bool
	LSPEnabled       bool
	Context          *config.Context
}

func (a *ApplicationOptions) GetLSPEnabled() bool {
	if a != nil {
		return a.LSPEnabled
	}
	return false
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
