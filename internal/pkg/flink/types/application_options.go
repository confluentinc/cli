package types

type ApplicationOptions struct {
	DefaultProperties map[string]string
	FlinkGatewayUrl   string
	UnsafeTrace       bool
	UserAgent         string
	EnvironmentId     string
	OrgResourceId     string
	KafkaClusterId    string
	ComputePoolId     string
	IdentityPoolId    string
	Verbose           bool
}

func (a *ApplicationOptions) GetDefaultProperties() map[string]string {
	if a != nil {
		return a.DefaultProperties
	}
	return map[string]string{}
}

func (a *ApplicationOptions) GetFlinkGatewayUrl() string {
	if a != nil {
		return a.FlinkGatewayUrl
	}
	return ""
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

func (a *ApplicationOptions) GetOrgResourceId() string {
	if a != nil {
		return a.OrgResourceId
	}
	return ""
}

func (a *ApplicationOptions) GetKafkaClusterId() string {
	if a != nil {
		return a.KafkaClusterId
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
