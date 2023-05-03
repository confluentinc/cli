package connect

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type installation struct {
	Type string
	Path string
	Use  string
}

type WorkerConfig struct {
	Path string
	Use  string
}

type listOut struct {
	Number      string `human:""`
	Path        string `human:"Path"`
	Description string `human:"Description"`
}

func getInstallation(cmd *cobra.Command, noPrompt bool) (*installation, error) {
	installations, err := findInstallationDirectories()
	if err != nil {
		return nil, err
	}

	if len(installations) == 0 {
		return nil, errors.NewErrorWithSuggestions("unable to detect Confluent Platform installation", "Pass the component directory and worker configuration files to the `--component-dir` and `--worker-configs` flags.")
	} else if noPrompt {
		return &installations[0], nil
	} else if len(installations) == 1 {
		output.Printf("Using the only available Confluent Platform installation at \"%s\".\n", installations[0].Path)
		return &installations[0], nil
	} else {
		list := output.NewList(cmd)
		for i, installation := range installations {
			list.Add(&listOut{
				Number:      strconv.Itoa(i + 1),
				Path:        installation.Path,
				Description: installation.Use,
			})
		}
		listStr, err := list.PrintString()
		if err != nil {
			return nil, err
		}

		f := form.New(form.Field{
			ID:     "installation",
			Prompt: fmt.Sprintf("Enter the number corresponding to the Confluent Platform installation you would like to use.\n%sTo cancel, press Ctrl-C", listStr),
			Regex:  `^\d$`,
		})
		if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
			return nil, err
		}
		choice, err := strconv.Atoi(f.Responses["installation"].(string))
		if err != nil || choice < 1 || choice > len(installations) {
			return nil, errors.Errorf("Number must be in the range %d to %d (inclusive)", 1, len(installations))
		}
		return &installations[choice-1], nil
	}
}

func findInstallationDirectories() ([]installation, error) {
	// Check in descending order of precedence:
	//   - $CONFLUENT_HOME
	//   - current directory
	//   - standard rpm/deb
	//   - based on the client

	var result []installation
	hasPackageInstallation := utils.DoesPathExist(filepath.FromSlash("/usr/bin/connect-distributed"))

	// $CONFLUENT_HOME
	confluentHome := os.Getenv("CONFLUENT_HOME")
	if confluentHome != "" && hasArchiveInstallation(confluentHome) {
		ins := installation{
			Type: archiveInstallation,
			Path: confluentHome,
			Use:  "$CONFLUENT_HOME",
		}
		result = append(result, ins)
	}

	// current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, errors.Errorf("unable to determine current working directory: %v", err)
	}
	if hasArchiveInstallation(currentDirectory) {
		ins := installation{
			Type: archiveInstallation,
			Path: currentDirectory,
			Use:  "Current Directory",
		}
		result = append(result, ins)
	}

	// standard rpm/deb
	if hasPackageInstallation {
		ins := installation{
			Type: packageInstallation,
			Path: filepath.FromSlash("/"),
			Use:  "Installed RPM/DEB Package",
		}
		result = append(result, ins)
	}

	// based on the client
	cliPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("unable to determine path to CLI: %v", err)
	}
	cliDirectory := filepath.Dir(cliPath)
	cliUse := "CLI Installation Directory"
	if hasArchiveInstallation(cliDirectory) {
		ins := installation{
			Type: archiveInstallation,
			Path: cliDirectory,
			Use:  cliUse,
		}
		result = append(result, ins)
	} else if filepath.ToSlash(cliDirectory) == "/usr/bin" && hasPackageInstallation {
		ins := installation{
			Type: packageInstallation,
			Path: filepath.FromSlash("/"),
			Use:  cliUse,
		}
		result = append(result, ins)
	}

	return result, nil
}

func hasArchiveInstallation(dir string) bool {
	if filepath.ToSlash(dir) == "/usr" {
		return false
	}

	return utils.DoesPathExist(filepath.Join(dir, filepath.FromSlash("share/java/confluent-common")))
}

func chooseComponentDir(ins *installation, noPrompt bool) (string, error) {
	var defaultComponentDir string
	switch ins.Type {
	case archiveInstallation:
		defaultComponentDir = filepath.Join(ins.Path, "share/confluent-hub-components")
	case packageInstallation:
		defaultComponentDir = "/usr/share/confluent-hub-components"
	default:
		return "", errors.Errorf("unexpected installation type: %s", ins.Type)
	}

	if noPrompt {
		return defaultComponentDir, nil
	}

	f := form.New(form.Field{
		ID:        "confirm",
		Prompt:    fmt.Sprintf("Do you want to install this plugin into %s?", defaultComponentDir),
		IsYesOrNo: true,
	})
	if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
		return "", err
	}
	if f.Responses["confirm"].(bool) {
		return defaultComponentDir, nil
	}

	f = form.New(form.Field{
		ID:     "directory",
		Prompt: "Specify installation directory",
	})
	if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
		return "", err
	}

	inputDir := f.Responses["directory"].(string)
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return "", errors.Errorf(`failed to determine absolute path to component directory "%s": %v`, inputDir, err)
	}
	if !utils.DoesPathExist(inputDir) {
		return "", errors.Errorf(`component directory "%s" does not exist`, inputDir)
	}

	return inputDir, nil
}

func standardWorkerConfigLocations(ins *installation) ([]WorkerConfig, error) {
	workerConfigLocations := []string{
		"/etc/kafka/connect-distributed.properties",
		"/etc/kafka/connect-standalone.properties",
		"/etc/schema-registry/connect-avro-distributed.properties",
		"/etc/schema-registry/connect-avro-standalone.properties",
	}
	switch ins.Type {
	case archiveInstallation:
		var result []WorkerConfig
		for _, workerConfigLocation := range workerConfigLocations {
			workerConfigPath := filepath.Join(ins.Path, filepath.FromSlash(workerConfigLocation))
			result = append(result, WorkerConfig{Path: workerConfigPath, Use: "Standard"})
		}
		confluentCurrentDir := os.Getenv("CONFLUENT_CURRENT")
		if confluentCurrentDir == "" {
			confluentCurrentDir = os.Getenv("TMPDIR")
		}
		confluentCurrentFile := filepath.Join(confluentCurrentDir, "confluent.current")
		if utils.DoesPathExist(confluentCurrentFile) {
			confluentCurrentContent, err := os.ReadFile(confluentCurrentFile)
			if err != nil {
				return nil, errors.Errorf("failed to read possible $CONFLUENT_CURRENT file %s: %v", confluentCurrentFile, err)
			}
			confluentCurrentLines := strings.SplitN(string(confluentCurrentContent), "\n", 3)
			if len(confluentCurrentLines) == 1 {
				connectCurrentConfigFile := filepath.Join(confluentCurrentLines[0], "/connect/connect.properties")
				result = append(result, WorkerConfig{Path: connectCurrentConfigFile, Use: "$CONFLUENT_CURRENT"})
			}
		}
		return result, nil
	case packageInstallation:
		var result []WorkerConfig
		for _, workerConfigLocation := range workerConfigLocations {
			result = append(result, WorkerConfig{Path: filepath.FromSlash(workerConfigLocation), Use: "Standard"})
		}
		return result, nil
	default:
		return nil, errors.New(fmt.Sprintf("unexpected installation type: %s", ins.Type))
	}
}

func runningWorkerConfigLocations() ([]WorkerConfig, error) {
	re := regexp.MustCompile(`org\.apache\.kafka\.connect\.cli\.Connect(Distributed|Standalone)`)

	command := `ps ax |
					grep -E '` + re.String() + `'|
					grep -v grep;
				test ${PIPESTATUS[0]} -eq 0`
	out, err := exec.Command("/bin/bash", "-c", command).Output()
	if err != nil {
		return nil, errors.Errorf("failed to run shell command to locate running Connect worker processes: %v", err)
	}

	var result []WorkerConfig
	for _, line := range strings.Split(string(out), "\n") {
		reachedArgs := false
		var pid string
		for i, word := range strings.Split(line, " ") {
			if i == 0 {
				pid = word
				continue
			}

			if re.MatchString(word) {
				reachedArgs = true
				continue
			}

			if reachedArgs && word != "-daemon" {
				// TODO: This doesn't work on workers that were started with relative paths to their config files
				//		 unless the CLI is run in the same directory that the Connect worker was started in
				result = append(result, WorkerConfig{Path: word, Use: "Used by Connect process with PID " + pid})
				break
			}
		}
	}
	return result, nil
}

func chooseWorkerConfigs(cmd *cobra.Command, ins *installation, noPrompt bool) ([]string, error) {
	var workerConfigs []WorkerConfig

	if standardWorkerConfigs, err := standardWorkerConfigLocations(ins); err != nil {
		return nil, errors.Errorf("could not infer possible worker config locations from standard candidates: %v", err)
	} else {
		for _, workerConfig := range standardWorkerConfigs {
			if utils.DoesPathExist(workerConfig.Path) {
				workerConfigs = append(workerConfigs, workerConfig)
			}
		}
	}

	if runningWorkerConfigs, err := runningWorkerConfigLocations(); err != nil {
		return nil, errors.Errorf("could not infer possible worker config locations from running processes: %v", err)
	} else {
		for _, workerConfig := range runningWorkerConfigs {
			if utils.DoesPathExist(workerConfig.Path) {
				workerConfigs = append(workerConfigs, workerConfig)
			}
		}
	}

	var filteredWorkerConfigs []WorkerConfig
	if noPrompt || len(workerConfigs) == 0 {
		filteredWorkerConfigs = workerConfigs
	} else {
		output.Println("Detected the following worker configs:")
		list := output.NewList(cmd)
		for i, workerConfig := range workerConfigs {
			list.Add(&listOut{
				Number:      strconv.Itoa(i + 1),
				Path:        workerConfig.Path,
				Description: workerConfig.Use,
			})
		}
		if err := list.Print(); err != nil {
			return nil, err
		}

		f := form.New(form.Field{
			ID:        "confirm",
			Prompt:    "Do you want to update all detected configs?",
			IsYesOrNo: true,
		})
		if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
			return nil, err
		}
		if f.Responses["confirm"].(bool) {
			filteredWorkerConfigs = workerConfigs
		} else {
			for i, workerConfig := range workerConfigs {
				f := form.New(form.Field{
					ID:        "confirm",
					Prompt:    fmt.Sprintf("Do you want to update config %d?", i+1),
					IsYesOrNo: true,
				})
				if err := f.Prompt(form.NewPrompt(os.Stdin)); err != nil {
					return nil, err
				}
				if f.Responses["confirm"].(bool) {
					filteredWorkerConfigs = append(filteredWorkerConfigs, workerConfig)
				}
			}
		}
	}

	var result []string
	for _, workerConfig := range filteredWorkerConfigs {
		result = append(result, workerConfig.Path)
	}
	return result, nil
}

func updateWorkerConfig(componentDir string, workerConfigPath string, dryRun bool) error {
	pluginPathProperty := "plugin.path"

	workerConfig, err := properties.LoadFile(workerConfigPath, properties.UTF8)
	if err != nil {
		return errors.Errorf("failed to parse worker config file %s: %v", workerConfigPath, err)
	}
	pluginPath := workerConfig.GetString(pluginPathProperty, "")
	pluginPathElements := regexp.MustCompile(" *, *").Split(pluginPath, -1)
	for _, pluginPathElement := range pluginPathElements {
		if pluginPathElement == componentDir {
			// Component directory is already included in the worker's plugin.path
			// No further action required on our part
			return nil
		}
	}
	newPluginPath := strings.Join(append(pluginPathElements, componentDir), ", ")
	if _, _, err = workerConfig.Set(pluginPathProperty, newPluginPath); err != nil {
		return errors.Errorf("failed to update %s property to %s for worker config %s: %v", pluginPathProperty, newPluginPath, workerConfigPath, err)
	}
	fileInfo, err := os.Stat(workerConfigPath)
	if err != nil {
		return err
	}
	if dryRun {
		fmt.Printf("Skipping update of worker config %s as part of dry run mode\n", workerConfigPath)
		return nil
	}
	workerConfigFile, err := os.OpenFile(workerConfigPath, os.O_TRUNC|os.O_RDWR, fileInfo.Mode())
	if err != nil {
		return errors.Errorf("failed to open worker config file %s before updating with new %s value %s: %v", workerConfigPath, pluginPathProperty, newPluginPath, err)
	}
	defer workerConfigFile.Close()
	// NOTE: This currently changes the comment spacing and removes empty lines
	if _, err = workerConfig.WriteFormattedComment(workerConfigFile, properties.UTF8); err != nil {
		return errors.Errorf("failed to update worker config file %s with new %s value %s: %v", workerConfigPath, pluginPathProperty, newPluginPath, err)
	}
	return nil
}
