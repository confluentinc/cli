package plugin

import (
	"github.com/hashicorp/go-version"
)

type BashPluginInstaller struct {
	Name          string
	RepositoryDir string
	InstallDir    string
}

func (b *BashPluginInstaller) CheckVersion(_ *version.Version) error {
	return nil
}

func (b *BashPluginInstaller) Install() error {
	return installSimplePlugin(b.Name, b.RepositoryDir, b.InstallDir, "bash")
}
