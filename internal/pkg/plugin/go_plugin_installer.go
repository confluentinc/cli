package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
)

const goVersionPattern = `^go[1-9][0-9]*\.[0-9]+(\.[1-9][0-9]*)?$`

type GoPluginInstaller struct {
	Name string
}

func (g *GoPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(programNotFoundErrorMsg, "go"), programNotFoundSuggestions)
	}

	re := regexp.MustCompile(goVersionPattern)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				return errors.Errorf(unableToParseVersionErrorMsg, "go")
			}
			if installedVer.LessThan(ver) {
				return errors.Errorf(insufficientVersionErrorMsg, "go", installedVer, ver)
			}
		}
	}

	return nil
}

func (g *GoPluginInstaller) Install() error {
	packageName := fmt.Sprintf("github.com/confluentinc/cli-plugins/%s@latest", g.Name)
	installCmd := exec.NewCommand("go", "install", packageName)

	if _, err := installCmd.Output(); err != nil {
		return errors.Wrap(err, "failed to run `go install`")
	}

	return nil
}
