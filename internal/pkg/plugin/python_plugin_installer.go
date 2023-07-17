package plugin

import (
	"fmt"
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

func (p *PythonPluginInstaller) IsVersion(word string) bool {
	re := regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.(0|[1-9][0-9]*)$`)

	return re.MatchString(word)
}

func (p *PythonPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("python", "--version")
	version3Cmd := exec.NewCommand("python3", "--version")

	versionSegments := ver.Segments()
	if len(versionSegments) == 0 {
		return errors.New(errors.NoVersionFoundErrorMsg)
	}
	majorVer := versionSegments[0]

	var out []byte
	var err error
	if majorVer == 3 {
		out, err = version3Cmd.Output()
	}
	if err != nil || majorVer != 3 {
		out, err = versionCmd.Output()
	}
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(programNotFoundErrorMsg, "python"), programNotFoundSuggestions)
	}

	for _, word := range strings.Split(string(out), " ") {
		if p.IsVersion(word) {
			installedVer, err := version.NewVersion(strings.TrimSpace(word))
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
