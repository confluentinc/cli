// Code generated by mocker. DO NOT EDIT.
// github.com/travisjeffery/mocker
// Source: pkg/auth/login_credentials_manager.go

package mock

import (
	sync "sync"

	github_com_confluentinc_ccloud_sdk_go_v1_public "github.com/confluentinc/ccloud-sdk-go-v1-public"
	github_com_confluentinc_cli_v4_pkg_auth "github.com/confluentinc/cli/v4/pkg/auth"
	github_com_confluentinc_cli_v4_pkg_config "github.com/confluentinc/cli/v4/pkg/config"
)

// LoginCredentialsManager is a mock of LoginCredentialsManager interface
type LoginCredentialsManager struct {
	lockGetCloudCredentialsFromEnvVar sync.Mutex
	GetCloudCredentialsFromEnvVarFunc func(arg0 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremCredentialsFromEnvVar sync.Mutex
	GetOnPremCredentialsFromEnvVarFunc func() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetCredentialsFromConfig sync.Mutex
	GetCredentialsFromConfigFunc func(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config, arg1 github_com_confluentinc_cli_v4_pkg_config.MachineParams) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetCredentialsFromKeychain sync.Mutex
	GetCredentialsFromKeychainFunc func(arg0 bool, arg1, arg2 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremSsoCredentials sync.Mutex
	GetOnPremSsoCredentialsFunc func(url, caCertPath, clientCertPath, clientKeyPath string, unsafeTrace bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremSsoCredentialsFromConfig sync.Mutex
	GetOnPremSsoCredentialsFromConfigFunc func(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config, arg1 bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremCertOnlyCredentials sync.Mutex
	GetOnPremCertOnlyCredentialsFunc func(certificateOnly bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetCloudCredentialsFromPrompt sync.Mutex
	GetCloudCredentialsFromPromptFunc func(arg0 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremCredentialsFromPrompt sync.Mutex
	GetOnPremCredentialsFromPromptFunc func() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetPrerunCredentialsFromConfig sync.Mutex
	GetPrerunCredentialsFromConfigFunc func(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockGetOnPremPrerunCredentialsFromEnvVar sync.Mutex
	GetOnPremPrerunCredentialsFromEnvVarFunc func() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error)

	lockSetCloudClient sync.Mutex
	SetCloudClientFunc func(arg0 *github_com_confluentinc_ccloud_sdk_go_v1_public.Client)

	calls struct {
		GetCloudCredentialsFromEnvVar []struct {
			Arg0 string
		}
		GetOnPremCredentialsFromEnvVar []struct {
		}
		GetCredentialsFromConfig []struct {
			Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
			Arg1 github_com_confluentinc_cli_v4_pkg_config.MachineParams
		}
		GetCredentialsFromKeychain []struct {
			Arg0 bool
			Arg1 string
			Arg2 string
		}
		GetOnPremSsoCredentials []struct {
			Url            string
			CaCertPath     string
			ClientCertPath string
			ClientKeyPath  string
			UnsafeTrace    bool
		}
		GetOnPremSsoCredentialsFromConfig []struct {
			Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
			Arg1 bool
		}
		GetOnPremCertOnlyCredentials []struct {
			CertificateOnly bool
		}
		GetCloudCredentialsFromPrompt []struct {
			Arg0 string
		}
		GetOnPremCredentialsFromPrompt []struct {
		}
		GetPrerunCredentialsFromConfig []struct {
			Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
		}
		GetOnPremPrerunCredentialsFromEnvVar []struct {
		}
		SetCloudClient []struct {
			Arg0 *github_com_confluentinc_ccloud_sdk_go_v1_public.Client
		}
	}
}

// GetCloudCredentialsFromEnvVar mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetCloudCredentialsFromEnvVar(arg0 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetCloudCredentialsFromEnvVar.Lock()
	defer m.lockGetCloudCredentialsFromEnvVar.Unlock()

	if m.GetCloudCredentialsFromEnvVarFunc == nil {
		panic("mocker: LoginCredentialsManager.GetCloudCredentialsFromEnvVarFunc is nil but LoginCredentialsManager.GetCloudCredentialsFromEnvVar was called.")
	}

	call := struct {
		Arg0 string
	}{
		Arg0: arg0,
	}

	m.calls.GetCloudCredentialsFromEnvVar = append(m.calls.GetCloudCredentialsFromEnvVar, call)

	return m.GetCloudCredentialsFromEnvVarFunc(arg0)
}

// GetCloudCredentialsFromEnvVarCalled returns true if GetCloudCredentialsFromEnvVar was called at least once.
func (m *LoginCredentialsManager) GetCloudCredentialsFromEnvVarCalled() bool {
	m.lockGetCloudCredentialsFromEnvVar.Lock()
	defer m.lockGetCloudCredentialsFromEnvVar.Unlock()

	return len(m.calls.GetCloudCredentialsFromEnvVar) > 0
}

// GetCloudCredentialsFromEnvVarCalls returns the calls made to GetCloudCredentialsFromEnvVar.
func (m *LoginCredentialsManager) GetCloudCredentialsFromEnvVarCalls() []struct {
	Arg0 string
} {
	m.lockGetCloudCredentialsFromEnvVar.Lock()
	defer m.lockGetCloudCredentialsFromEnvVar.Unlock()

	return m.calls.GetCloudCredentialsFromEnvVar
}

// GetOnPremCredentialsFromEnvVar mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromEnvVar() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremCredentialsFromEnvVar.Unlock()

	if m.GetOnPremCredentialsFromEnvVarFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremCredentialsFromEnvVarFunc is nil but LoginCredentialsManager.GetOnPremCredentialsFromEnvVar was called.")
	}

	call := struct {
	}{}

	m.calls.GetOnPremCredentialsFromEnvVar = append(m.calls.GetOnPremCredentialsFromEnvVar, call)

	return m.GetOnPremCredentialsFromEnvVarFunc()
}

// GetOnPremCredentialsFromEnvVarCalled returns true if GetOnPremCredentialsFromEnvVar was called at least once.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromEnvVarCalled() bool {
	m.lockGetOnPremCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremCredentialsFromEnvVar.Unlock()

	return len(m.calls.GetOnPremCredentialsFromEnvVar) > 0
}

// GetOnPremCredentialsFromEnvVarCalls returns the calls made to GetOnPremCredentialsFromEnvVar.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromEnvVarCalls() []struct {
} {
	m.lockGetOnPremCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremCredentialsFromEnvVar.Unlock()

	return m.calls.GetOnPremCredentialsFromEnvVar
}

// GetCredentialsFromConfig mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetCredentialsFromConfig(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config, arg1 github_com_confluentinc_cli_v4_pkg_config.MachineParams) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetCredentialsFromConfig.Lock()
	defer m.lockGetCredentialsFromConfig.Unlock()

	if m.GetCredentialsFromConfigFunc == nil {
		panic("mocker: LoginCredentialsManager.GetCredentialsFromConfigFunc is nil but LoginCredentialsManager.GetCredentialsFromConfig was called.")
	}

	call := struct {
		Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
		Arg1 github_com_confluentinc_cli_v4_pkg_config.MachineParams
	}{
		Arg0: arg0,
		Arg1: arg1,
	}

	m.calls.GetCredentialsFromConfig = append(m.calls.GetCredentialsFromConfig, call)

	return m.GetCredentialsFromConfigFunc(arg0, arg1)
}

// GetCredentialsFromConfigCalled returns true if GetCredentialsFromConfig was called at least once.
func (m *LoginCredentialsManager) GetCredentialsFromConfigCalled() bool {
	m.lockGetCredentialsFromConfig.Lock()
	defer m.lockGetCredentialsFromConfig.Unlock()

	return len(m.calls.GetCredentialsFromConfig) > 0
}

// GetCredentialsFromConfigCalls returns the calls made to GetCredentialsFromConfig.
func (m *LoginCredentialsManager) GetCredentialsFromConfigCalls() []struct {
	Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
	Arg1 github_com_confluentinc_cli_v4_pkg_config.MachineParams
} {
	m.lockGetCredentialsFromConfig.Lock()
	defer m.lockGetCredentialsFromConfig.Unlock()

	return m.calls.GetCredentialsFromConfig
}

// GetCredentialsFromKeychain mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetCredentialsFromKeychain(arg0 bool, arg1, arg2 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetCredentialsFromKeychain.Lock()
	defer m.lockGetCredentialsFromKeychain.Unlock()

	if m.GetCredentialsFromKeychainFunc == nil {
		panic("mocker: LoginCredentialsManager.GetCredentialsFromKeychainFunc is nil but LoginCredentialsManager.GetCredentialsFromKeychain was called.")
	}

	call := struct {
		Arg0 bool
		Arg1 string
		Arg2 string
	}{
		Arg0: arg0,
		Arg1: arg1,
		Arg2: arg2,
	}

	m.calls.GetCredentialsFromKeychain = append(m.calls.GetCredentialsFromKeychain, call)

	return m.GetCredentialsFromKeychainFunc(arg0, arg1, arg2)
}

// GetCredentialsFromKeychainCalled returns true if GetCredentialsFromKeychain was called at least once.
func (m *LoginCredentialsManager) GetCredentialsFromKeychainCalled() bool {
	m.lockGetCredentialsFromKeychain.Lock()
	defer m.lockGetCredentialsFromKeychain.Unlock()

	return len(m.calls.GetCredentialsFromKeychain) > 0
}

// GetCredentialsFromKeychainCalls returns the calls made to GetCredentialsFromKeychain.
func (m *LoginCredentialsManager) GetCredentialsFromKeychainCalls() []struct {
	Arg0 bool
	Arg1 string
	Arg2 string
} {
	m.lockGetCredentialsFromKeychain.Lock()
	defer m.lockGetCredentialsFromKeychain.Unlock()

	return m.calls.GetCredentialsFromKeychain
}

// GetOnPremSsoCredentials mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremSsoCredentials(url, caCertPath, clientCertPath, clientKeyPath string, unsafeTrace bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremSsoCredentials.Lock()
	defer m.lockGetOnPremSsoCredentials.Unlock()

	if m.GetOnPremSsoCredentialsFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremSsoCredentialsFunc is nil but LoginCredentialsManager.GetOnPremSsoCredentials was called.")
	}

	call := struct {
		Url            string
		CaCertPath     string
		ClientCertPath string
		ClientKeyPath  string
		UnsafeTrace    bool
	}{
		Url:            url,
		CaCertPath:     caCertPath,
		ClientCertPath: clientCertPath,
		ClientKeyPath:  clientKeyPath,
		UnsafeTrace:    unsafeTrace,
	}

	m.calls.GetOnPremSsoCredentials = append(m.calls.GetOnPremSsoCredentials, call)

	return m.GetOnPremSsoCredentialsFunc(url, caCertPath, clientCertPath, clientKeyPath, unsafeTrace)
}

// GetOnPremSsoCredentialsCalled returns true if GetOnPremSsoCredentials was called at least once.
func (m *LoginCredentialsManager) GetOnPremSsoCredentialsCalled() bool {
	m.lockGetOnPremSsoCredentials.Lock()
	defer m.lockGetOnPremSsoCredentials.Unlock()

	return len(m.calls.GetOnPremSsoCredentials) > 0
}

// GetOnPremSsoCredentialsCalls returns the calls made to GetOnPremSsoCredentials.
func (m *LoginCredentialsManager) GetOnPremSsoCredentialsCalls() []struct {
	Url            string
	CaCertPath     string
	ClientCertPath string
	ClientKeyPath  string
	UnsafeTrace    bool
} {
	m.lockGetOnPremSsoCredentials.Lock()
	defer m.lockGetOnPremSsoCredentials.Unlock()

	return m.calls.GetOnPremSsoCredentials
}

// GetOnPremSsoCredentialsFromConfig mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremSsoCredentialsFromConfig(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config, arg1 bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremSsoCredentialsFromConfig.Lock()
	defer m.lockGetOnPremSsoCredentialsFromConfig.Unlock()

	if m.GetOnPremSsoCredentialsFromConfigFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremSsoCredentialsFromConfigFunc is nil but LoginCredentialsManager.GetOnPremSsoCredentialsFromConfig was called.")
	}

	call := struct {
		Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
		Arg1 bool
	}{
		Arg0: arg0,
		Arg1: arg1,
	}

	m.calls.GetOnPremSsoCredentialsFromConfig = append(m.calls.GetOnPremSsoCredentialsFromConfig, call)

	return m.GetOnPremSsoCredentialsFromConfigFunc(arg0, arg1)
}

// GetOnPremSsoCredentialsFromConfigCalled returns true if GetOnPremSsoCredentialsFromConfig was called at least once.
func (m *LoginCredentialsManager) GetOnPremSsoCredentialsFromConfigCalled() bool {
	m.lockGetOnPremSsoCredentialsFromConfig.Lock()
	defer m.lockGetOnPremSsoCredentialsFromConfig.Unlock()

	return len(m.calls.GetOnPremSsoCredentialsFromConfig) > 0
}

// GetOnPremSsoCredentialsFromConfigCalls returns the calls made to GetOnPremSsoCredentialsFromConfig.
func (m *LoginCredentialsManager) GetOnPremSsoCredentialsFromConfigCalls() []struct {
	Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
	Arg1 bool
} {
	m.lockGetOnPremSsoCredentialsFromConfig.Lock()
	defer m.lockGetOnPremSsoCredentialsFromConfig.Unlock()

	return m.calls.GetOnPremSsoCredentialsFromConfig
}

// GetOnPremCertOnlyCredentials mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremCertOnlyCredentials(certificateOnly bool) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremCertOnlyCredentials.Lock()
	defer m.lockGetOnPremCertOnlyCredentials.Unlock()

	if m.GetOnPremCertOnlyCredentialsFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremCertOnlyCredentialsFunc is nil but LoginCredentialsManager.GetOnPremCertOnlyCredentials was called.")
	}

	call := struct {
		CertificateOnly bool
	}{
		CertificateOnly: certificateOnly,
	}

	m.calls.GetOnPremCertOnlyCredentials = append(m.calls.GetOnPremCertOnlyCredentials, call)

	return m.GetOnPremCertOnlyCredentialsFunc(certificateOnly)
}

// GetOnPremCertOnlyCredentialsCalled returns true if GetOnPremCertOnlyCredentials was called at least once.
func (m *LoginCredentialsManager) GetOnPremCertOnlyCredentialsCalled() bool {
	m.lockGetOnPremCertOnlyCredentials.Lock()
	defer m.lockGetOnPremCertOnlyCredentials.Unlock()

	return len(m.calls.GetOnPremCertOnlyCredentials) > 0
}

// GetOnPremCertOnlyCredentialsCalls returns the calls made to GetOnPremCertOnlyCredentials.
func (m *LoginCredentialsManager) GetOnPremCertOnlyCredentialsCalls() []struct {
	CertificateOnly bool
} {
	m.lockGetOnPremCertOnlyCredentials.Lock()
	defer m.lockGetOnPremCertOnlyCredentials.Unlock()

	return m.calls.GetOnPremCertOnlyCredentials
}

// GetCloudCredentialsFromPrompt mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetCloudCredentialsFromPrompt(arg0 string) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetCloudCredentialsFromPrompt.Lock()
	defer m.lockGetCloudCredentialsFromPrompt.Unlock()

	if m.GetCloudCredentialsFromPromptFunc == nil {
		panic("mocker: LoginCredentialsManager.GetCloudCredentialsFromPromptFunc is nil but LoginCredentialsManager.GetCloudCredentialsFromPrompt was called.")
	}

	call := struct {
		Arg0 string
	}{
		Arg0: arg0,
	}

	m.calls.GetCloudCredentialsFromPrompt = append(m.calls.GetCloudCredentialsFromPrompt, call)

	return m.GetCloudCredentialsFromPromptFunc(arg0)
}

// GetCloudCredentialsFromPromptCalled returns true if GetCloudCredentialsFromPrompt was called at least once.
func (m *LoginCredentialsManager) GetCloudCredentialsFromPromptCalled() bool {
	m.lockGetCloudCredentialsFromPrompt.Lock()
	defer m.lockGetCloudCredentialsFromPrompt.Unlock()

	return len(m.calls.GetCloudCredentialsFromPrompt) > 0
}

// GetCloudCredentialsFromPromptCalls returns the calls made to GetCloudCredentialsFromPrompt.
func (m *LoginCredentialsManager) GetCloudCredentialsFromPromptCalls() []struct {
	Arg0 string
} {
	m.lockGetCloudCredentialsFromPrompt.Lock()
	defer m.lockGetCloudCredentialsFromPrompt.Unlock()

	return m.calls.GetCloudCredentialsFromPrompt
}

// GetOnPremCredentialsFromPrompt mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromPrompt() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremCredentialsFromPrompt.Lock()
	defer m.lockGetOnPremCredentialsFromPrompt.Unlock()

	if m.GetOnPremCredentialsFromPromptFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremCredentialsFromPromptFunc is nil but LoginCredentialsManager.GetOnPremCredentialsFromPrompt was called.")
	}

	call := struct {
	}{}

	m.calls.GetOnPremCredentialsFromPrompt = append(m.calls.GetOnPremCredentialsFromPrompt, call)

	return m.GetOnPremCredentialsFromPromptFunc()
}

// GetOnPremCredentialsFromPromptCalled returns true if GetOnPremCredentialsFromPrompt was called at least once.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromPromptCalled() bool {
	m.lockGetOnPremCredentialsFromPrompt.Lock()
	defer m.lockGetOnPremCredentialsFromPrompt.Unlock()

	return len(m.calls.GetOnPremCredentialsFromPrompt) > 0
}

// GetOnPremCredentialsFromPromptCalls returns the calls made to GetOnPremCredentialsFromPrompt.
func (m *LoginCredentialsManager) GetOnPremCredentialsFromPromptCalls() []struct {
} {
	m.lockGetOnPremCredentialsFromPrompt.Lock()
	defer m.lockGetOnPremCredentialsFromPrompt.Unlock()

	return m.calls.GetOnPremCredentialsFromPrompt
}

// GetPrerunCredentialsFromConfig mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetPrerunCredentialsFromConfig(arg0 *github_com_confluentinc_cli_v4_pkg_config.Config) func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetPrerunCredentialsFromConfig.Lock()
	defer m.lockGetPrerunCredentialsFromConfig.Unlock()

	if m.GetPrerunCredentialsFromConfigFunc == nil {
		panic("mocker: LoginCredentialsManager.GetPrerunCredentialsFromConfigFunc is nil but LoginCredentialsManager.GetPrerunCredentialsFromConfig was called.")
	}

	call := struct {
		Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
	}{
		Arg0: arg0,
	}

	m.calls.GetPrerunCredentialsFromConfig = append(m.calls.GetPrerunCredentialsFromConfig, call)

	return m.GetPrerunCredentialsFromConfigFunc(arg0)
}

// GetPrerunCredentialsFromConfigCalled returns true if GetPrerunCredentialsFromConfig was called at least once.
func (m *LoginCredentialsManager) GetPrerunCredentialsFromConfigCalled() bool {
	m.lockGetPrerunCredentialsFromConfig.Lock()
	defer m.lockGetPrerunCredentialsFromConfig.Unlock()

	return len(m.calls.GetPrerunCredentialsFromConfig) > 0
}

// GetPrerunCredentialsFromConfigCalls returns the calls made to GetPrerunCredentialsFromConfig.
func (m *LoginCredentialsManager) GetPrerunCredentialsFromConfigCalls() []struct {
	Arg0 *github_com_confluentinc_cli_v4_pkg_config.Config
} {
	m.lockGetPrerunCredentialsFromConfig.Lock()
	defer m.lockGetPrerunCredentialsFromConfig.Unlock()

	return m.calls.GetPrerunCredentialsFromConfig
}

// GetOnPremPrerunCredentialsFromEnvVar mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) GetOnPremPrerunCredentialsFromEnvVar() func() (*github_com_confluentinc_cli_v4_pkg_auth.Credentials, error) {
	m.lockGetOnPremPrerunCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremPrerunCredentialsFromEnvVar.Unlock()

	if m.GetOnPremPrerunCredentialsFromEnvVarFunc == nil {
		panic("mocker: LoginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVarFunc is nil but LoginCredentialsManager.GetOnPremPrerunCredentialsFromEnvVar was called.")
	}

	call := struct {
	}{}

	m.calls.GetOnPremPrerunCredentialsFromEnvVar = append(m.calls.GetOnPremPrerunCredentialsFromEnvVar, call)

	return m.GetOnPremPrerunCredentialsFromEnvVarFunc()
}

// GetOnPremPrerunCredentialsFromEnvVarCalled returns true if GetOnPremPrerunCredentialsFromEnvVar was called at least once.
func (m *LoginCredentialsManager) GetOnPremPrerunCredentialsFromEnvVarCalled() bool {
	m.lockGetOnPremPrerunCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremPrerunCredentialsFromEnvVar.Unlock()

	return len(m.calls.GetOnPremPrerunCredentialsFromEnvVar) > 0
}

// GetOnPremPrerunCredentialsFromEnvVarCalls returns the calls made to GetOnPremPrerunCredentialsFromEnvVar.
func (m *LoginCredentialsManager) GetOnPremPrerunCredentialsFromEnvVarCalls() []struct {
} {
	m.lockGetOnPremPrerunCredentialsFromEnvVar.Lock()
	defer m.lockGetOnPremPrerunCredentialsFromEnvVar.Unlock()

	return m.calls.GetOnPremPrerunCredentialsFromEnvVar
}

// SetCloudClient mocks base method by wrapping the associated func.
func (m *LoginCredentialsManager) SetCloudClient(arg0 *github_com_confluentinc_ccloud_sdk_go_v1_public.Client) {
	m.lockSetCloudClient.Lock()
	defer m.lockSetCloudClient.Unlock()

	if m.SetCloudClientFunc == nil {
		panic("mocker: LoginCredentialsManager.SetCloudClientFunc is nil but LoginCredentialsManager.SetCloudClient was called.")
	}

	call := struct {
		Arg0 *github_com_confluentinc_ccloud_sdk_go_v1_public.Client
	}{
		Arg0: arg0,
	}

	m.calls.SetCloudClient = append(m.calls.SetCloudClient, call)

	m.SetCloudClientFunc(arg0)
}

// SetCloudClientCalled returns true if SetCloudClient was called at least once.
func (m *LoginCredentialsManager) SetCloudClientCalled() bool {
	m.lockSetCloudClient.Lock()
	defer m.lockSetCloudClient.Unlock()

	return len(m.calls.SetCloudClient) > 0
}

// SetCloudClientCalls returns the calls made to SetCloudClient.
func (m *LoginCredentialsManager) SetCloudClientCalls() []struct {
	Arg0 *github_com_confluentinc_ccloud_sdk_go_v1_public.Client
} {
	m.lockSetCloudClient.Lock()
	defer m.lockSetCloudClient.Unlock()

	return m.calls.SetCloudClient
}

// Reset resets the calls made to the mocked methods.
func (m *LoginCredentialsManager) Reset() {
	m.lockGetCloudCredentialsFromEnvVar.Lock()
	m.calls.GetCloudCredentialsFromEnvVar = nil
	m.lockGetCloudCredentialsFromEnvVar.Unlock()
	m.lockGetOnPremCredentialsFromEnvVar.Lock()
	m.calls.GetOnPremCredentialsFromEnvVar = nil
	m.lockGetOnPremCredentialsFromEnvVar.Unlock()
	m.lockGetCredentialsFromConfig.Lock()
	m.calls.GetCredentialsFromConfig = nil
	m.lockGetCredentialsFromConfig.Unlock()
	m.lockGetCredentialsFromKeychain.Lock()
	m.calls.GetCredentialsFromKeychain = nil
	m.lockGetCredentialsFromKeychain.Unlock()
	m.lockGetOnPremSsoCredentials.Lock()
	m.calls.GetOnPremSsoCredentials = nil
	m.lockGetOnPremSsoCredentials.Unlock()
	m.lockGetOnPremSsoCredentialsFromConfig.Lock()
	m.calls.GetOnPremSsoCredentialsFromConfig = nil
	m.lockGetOnPremSsoCredentialsFromConfig.Unlock()
	m.lockGetOnPremCertOnlyCredentials.Lock()
	m.calls.GetOnPremCertOnlyCredentials = nil
	m.lockGetOnPremCertOnlyCredentials.Unlock()
	m.lockGetCloudCredentialsFromPrompt.Lock()
	m.calls.GetCloudCredentialsFromPrompt = nil
	m.lockGetCloudCredentialsFromPrompt.Unlock()
	m.lockGetOnPremCredentialsFromPrompt.Lock()
	m.calls.GetOnPremCredentialsFromPrompt = nil
	m.lockGetOnPremCredentialsFromPrompt.Unlock()
	m.lockGetPrerunCredentialsFromConfig.Lock()
	m.calls.GetPrerunCredentialsFromConfig = nil
	m.lockGetPrerunCredentialsFromConfig.Unlock()
	m.lockGetOnPremPrerunCredentialsFromEnvVar.Lock()
	m.calls.GetOnPremPrerunCredentialsFromEnvVar = nil
	m.lockGetOnPremPrerunCredentialsFromEnvVar.Unlock()
	m.lockSetCloudClient.Lock()
	m.calls.SetCloudClient = nil
	m.lockSetCloudClient.Unlock()
}
