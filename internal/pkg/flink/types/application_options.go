package types

import (
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
)

type ApplicationOptions struct {
	UnsafeTrace     bool
	UserAgent       string
	EnvironmentId   string
	EnvironmentName string
	OrgResourceId   string
	Database        string
	ComputePoolId   string
	IdentityPoolId  string
	Verbose         bool
	Context         *dynamicconfig.DynamicContext
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

func (a *ApplicationOptions) GetOrgResourceId() string {
	if a != nil {
		return a.OrgResourceId
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

func (a *ApplicationOptions) GetIdentityPoolId() string {
	if a != nil {
		return a.IdentityPoolId
	}
	return ""
}
func (a *ApplicationOptions) GetVerbose() bool {
	if a != nil {
		return a.Verbose
	}
	return false
}

func (a *ApplicationOptions) GetContext() *dynamicconfig.DynamicContext {
	if a != nil {
		return a.Context
	}
	return nil
}
