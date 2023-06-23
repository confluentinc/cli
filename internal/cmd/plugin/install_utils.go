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

const (
	programNotFoundMsg      = "[WARN] Unable to find %s. Check that it is installed in a directory in your $PATH.\n"
	unableToParseVersionMsg = "[WARN] Unable to parse %s version.\n"
	insufficientVersionMsg  = "[WARN] Installed %s version %s is less than the required version %s.\n"
)

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

func checkPythonVersion(ver *version.Version) {
	versionCmd := exec.NewCommand("python", "--version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "Python")
		return
	}

	re := regexp.MustCompile(`[1-9]\.[0-9]+[\.0-9]*`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.Trim(word, " \n"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "Python")
				return
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf(insufficientVersionMsg, "Python", installedVer, ver)
				return
			}
		}
	}
}

func checkGoVersion(ver *version.Version) {
	versionCmd := exec.NewCommand("go", "version")

	out, err := versionCmd.Output()
	if err != nil {
		output.ErrPrintf(programNotFoundMsg, "Go")
		return
	}

	re := regexp.MustCompile(`go[1-9]\.[0-9]+[\.0-9]*`)
	for _, word := range strings.Split(string(out), " ") {
		if re.MatchString(word) {
			installedVer, err := version.NewVersion(strings.TrimPrefix(word, "go"))
			if err != nil {
				output.ErrPrintf(unableToParseVersionMsg, "Go")
				return
			}
			if installedVer.LessThan(ver) {
				output.ErrPrintf(insufficientVersionMsg, "Go", installedVer, ver)
				return
			}
		}
	}
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
