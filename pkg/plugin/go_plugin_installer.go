package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/exec"
)

type GoPluginInstaller struct {
	Name string
}

func (g *GoPluginInstaller) IsVersion(word string) bool {
	re := regexp.MustCompile(`^go[1-9][0-9]*\.[0-9]+(\.[1-9][0-9]*)?$`)

	return re.MatchString(word)
}

func (g *GoPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(programNotFoundErrorMsg, "go"),
			programNotFoundSuggestions,
		)
	}

	for _, word := range strings.Split(string(out), " ") {
		if g.IsVersion(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				return fmt.Errorf(unableToParseVersionErrorMsg, "go")
			}
			if installedVer.LessThan(ver) {
				return fmt.Errorf(insufficientVersionErrorMsg, "go", installedVer, ver)
			}
		}
	}

	return nil
}

func (g *GoPluginInstaller) Install() error {
	packageName := fmt.Sprintf("github.com/confluentinc/cli-plugins/%s@latest", g.Name)
	installCmd := exec.NewCommand("go", "install", packageName)

	if _, err := installCmd.Output(); err != nil {
		return fmt.Errorf("failed to run `go install`: %w", err)
	}

	return nil
}
