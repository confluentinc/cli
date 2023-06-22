package plugin

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var supportedLanguages = []string{"python", "go"}

func getLanguage(manifest *Manifest) (string, *version.Version, error) {
	if manifest == nil {
		return "", nil, nil
	}

	if manifest.Dependencies == "" {
		return "", nil, nil
	}

	dependencySlice := strings.Split(strings.ToLower(manifest.Dependencies), " ")
	if len(dependencySlice) == 1 {
		return dependencySlice[0], nil, nil
	}

	ver, err := version.NewVersion(dependencySlice[1])
	if err != nil {
		return dependencySlice[0], nil, nil
	}

	return dependencySlice[0], ver, nil
}

func checkPythonVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("python", "--version")

	_, err := versionCmd.Output()
	if err != nil {
		return err
	}

	return nil
}

func checkGoVersion(ver *version.Version) error {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		return err
	}

	re := regexp.MustCompile(`go[1-9]\.[0-9]+[\.0-9]*`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				return err
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf("[WARN] Installed Go version %s is less than the required version %s.\n", installedVer, ver)
			}

			break
		}
	}

	return nil
}

func installPythonPlugin(name, repoDir, installDir string) error {
	pluginDir := fmt.Sprintf("%s/%s", repoDir, name)
	files, err := os.ReadDir(pluginDir)
	if err != nil {
		return err
	}

	found := false
	for _, file := range files {
		if strings.HasPrefix(file.Name(), "confluent-") && strings.HasSuffix(file.Name(), ".py") {
			found = true

			fileData, err := os.ReadFile(fmt.Sprintf("%s/%s", pluginDir, file.Name()))
			if err != nil {
				return err
			}

			if err := os.WriteFile(fmt.Sprintf("%s/%s", installDir, file.Name()), fileData, 0755); err != nil {
				return err
			}
		}
	}

	if !found {
		return errors.Errorf("unable to find .py file for plugin %s", name)
	}
	return nil
}

func installGoPlugin(name string) error {
	packageName := fmt.Sprintf("github.com/confluentinc/cli-plugins/%s@latest", name)
	installCmd := exec.NewCommand("go", "install", packageName)

	if _, err := installCmd.Output(); err != nil {
		return errors.Wrap(err, "failed to run go install command")
	}

	return nil
}
