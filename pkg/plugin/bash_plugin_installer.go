package plugin

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/exec"
)

type BashPluginInstaller struct {
	Id            string
	RepositoryDir string
	InstallDir    string
}

func (b *BashPluginInstaller) IsVersion(word string) bool {
	re := regexp.MustCompile(`^[1-9][0-9]*\.[0-9]+\.[0-9]+\([0-9]+\)-[a-z0-9]*$`)

	return re.MatchString(word)
}

func (b *BashPluginInstaller) CheckVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("bash", "--version")

	out, err := versionCmd.Output()
	if err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(programNotFoundErrorMsg, "bash"),
			programNotFoundSuggestions,
		)
	}

	for _, word := range strings.Split(string(out), " ") {
		if b.IsVersion(word) {
			parenthesisIdx := strings.Index(word, "(")
			installedVer, err := version.NewVersion(word[:parenthesisIdx])
			if err != nil {
				return fmt.Errorf(unableToParseVersionErrorMsg, "bash")
			}
			if installedVer.GreaterThan(ver) {
				return fmt.Errorf(insufficientVersionErrorMsg, "bash", installedVer, ver)
			}
		}
	}

	return nil
}

func (b *BashPluginInstaller) Install() error {
	return installSimplePlugin(b.Id, b.RepositoryDir, b.InstallDir, "bash")
}
