package plugin

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
)

type PythonPluginInstaller struct {
	Name          string
	RepositoryDir string
	InstallDir    string
}

func (p *PythonPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("python", "--version")
	version3Cmd := exec.NewCommand("python3", "--version")

	v3, _ := version.NewVersion("3.0.0")

	var out []byte
	var err error
	if ver.GreaterThanOrEqual(v3) {
		out, err = version3Cmd.Output()
	}
	if err != nil || ver.LessThan(v3) {
		out, err = versionCmd.Output()
	}
	if err != nil {
		return errors.Errorf(programNotFoundErrorMsg, "python")
	}

	re := regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.(0|[1-9][0-9]*)$`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.Trim(word, " \n"))
			if err != nil {
				return errors.Errorf(unableToParseVersionErrorMsg, "python")
			}
			if installedVer.LessThan(ver) {
				return errors.Errorf(insufficientVersionErrorMsg, "python", installedVer, ver)
			}
		}
	}

	return nil
}

func (p *PythonPluginInstaller) Install() error {
	return installSimplePlugin(p.Name, p.RepositoryDir, p.InstallDir, "python")
}
