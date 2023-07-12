package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type GoPluginInstaller struct {
	Name string
}

func (g *GoPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "go")
		return nil
	}

	re := regexp.MustCompile(`^go[1-9][0-9]*\.[0-9]+(\.[1-9][0-9]*)?$`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "go")
				return nil
			}
			if installedVer.LessThan(ver) {
				return errors.Errorf(insufficientVersionMsg, "go", installedVer, ver)
			}
		}
	}

	return nil
}

func (g *GoPluginInstaller) Install() error {
	packageName := fmt.Sprintf("github.com/confluentinc/cli-plugins/%s@latest", g.Name)
	installCmd := exec.NewCommand("go", "install", packageName)

	if _, err := installCmd.Output(); err != nil {
		return errors.Wrap(err, "failed to run go install command")
	}

	return nil
}
