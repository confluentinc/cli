package plugin

import (
	"github.com/hashicorp/go-version"
)

type ShellPluginInstaller struct {
	Name          string
	RepositoryDir string
	InstallDir    string
}

func (s *ShellPluginInstaller) CheckVersion(_ *version.Version) error {
	return nil
}

func (s *ShellPluginInstaller) Install() error {
	return installSimplePlugin(s.Name, s.RepositoryDir, s.InstallDir, "shell")
}
