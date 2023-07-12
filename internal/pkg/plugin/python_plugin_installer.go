package plugin

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/exec"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type PythonPluginInstaller struct {
	Name          string
	RepositoryDir string
	InstallDir    string
}

func (p *PythonPluginInstaller) CheckVersion(ver *version.Version) {
	versionCmd := exec.NewCommand("python", "--version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "python")
		return
	}

	re := regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.(0|[1-9][0-9]*)$`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.Trim(word, " \n"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "python")
				return
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf(insufficientVersionMsg, "python", installedVer, ver)
				return
			}
		}
	}
}

func (p *PythonPluginInstaller) Install() error {
	return installSimplePlugin(p.Name, p.RepositoryDir, p.InstallDir, "python")
}
