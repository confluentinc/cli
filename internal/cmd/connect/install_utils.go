package connect

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/exec"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/types"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type platformInstallation struct {
	Location platformLocation
	Use      string
}

type platformLocation struct {
	Type string
	Path string
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

func getConfluentPlatformInstallation(cmd *cobra.Command, prompt form.Prompt, force bool) (*platformInstallation, error) {
	if cmd.Flags().Changed("confluent-platform") {
		return getPlatformInstallationFromFlag(cmd)
	}

	installations, err := findInstallationDirectories()
	if err != nil {
		return nil, err
	}

	if len(installations) == 0 {
		return nil, errors.NewErrorWithSuggestions("unable to detect a Confluent Platform installation", "Pass the plugin directory and worker configuration files with `--plugin-directory` and `--worker-configs`.")
	} else if force {
		output.Printf("Using the Confluent Platform installation at \"%s\".\n", installations[0].Location.Path)
		return &installations[0], nil
	} else if len(installations) == 1 {
		output.Printf("Using the only available Confluent Platform installation at \"%s\".\n", installations[0].Location.Path)
		return &installations[0], nil
	} else {
		list := output.NewList(cmd)
		for i, installation := range installations {
			list.Add(&listOut{
				Number:      strconv.Itoa(i + 1),
				Path:        installation.Location.Path,
				Description: installation.Use,
			})
		}

		listStr, err := list.PrintString()
		if err != nil {
			return nil, err
		}

		promptMsg := "The plugin can be installed in any of the following Confluent Platform installations. Enter the number corresponding to the installation you would like to use:\n%sTo cancel, press Ctrl-C"
		f := form.New(form.Field{
			ID:     "installation",
			Prompt: fmt.Sprintf(promptMsg, listStr),
			Regex:  `^\d$`,
		})
		if err := f.Prompt(prompt); err != nil {
			return nil, err
		}
		choice, err := strconv.Atoi(f.Responses["installation"].(string))
		if err != nil || choice < 1 || choice > len(installations) {
			return nil, errors.Errorf("your choice must be in the range %d to %d (inclusive)", 1, len(installations))
		}
		return &installations[choice-1], nil
	}
}

func getPlatformInstallationFromFlag(cmd *cobra.Command) (*platformInstallation, error) {
	specifiedDirectory, err := cmd.Flags().GetString("confluent-platform")
	if err != nil {
		return nil, err
	}

	if !hasArchiveInstallation(specifiedDirectory) {
		return nil, errors.New("the directory specified with `--confluent-platform` does not correspond to a valid archive installation")
	}

	return &platformInstallation{
		Location: platformLocation{
			Type: "ARCHIVE",
			Path: specifiedDirectory,
		},
	}, nil
}

func findInstallationDirectories() ([]platformInstallation, error) {
	// Check in descending order of precedence:
	//   - $CONFLUENT_HOME
	//   - current directory
	//   - standard rpm/deb
	//   - based on the client

	var result []platformInstallation
	hasPackageInstallation := utils.DoesPathExist(filepath.FromSlash("/usr/bin/connect-distributed"))

	// $CONFLUENT_HOME
	confluentHome := os.Getenv("CONFLUENT_HOME")
	if confluentHome != "" && hasArchiveInstallation(confluentHome) {
		installation := platformInstallation{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: confluentHome,
			},
			Use: "$CONFLUENT_HOME",
		}
		result = append(result, installation)
	}

	// current directory
	currentDirectory, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine current working directory")
	}
	if hasArchiveInstallation(currentDirectory) {
		installation := platformInstallation{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: currentDirectory,
			},
			Use: "Current Directory",
		}
		result = append(result, installation)
	}

	// standard rpm/deb
	if hasPackageInstallation {
		installation := platformInstallation{
			Location: platformLocation{
				Type: "PACKAGE",
				Path: filepath.FromSlash("/"),
			},
			Use: "Installed RPM/DEB Package",
		}
		result = append(result, installation)
	}

	// based on the client
	cliPath, err := os.Executable()
	if err != nil {
		return nil, errors.Wrap(err, "unable to determine path to CLI")
	}
	cliDirectory := filepath.Dir(cliPath)
	cliUse := "CLI Installation Directory"
	if filepath.ToSlash(cliDirectory) == "/usr/bin" && hasPackageInstallation {
		installation := platformInstallation{
			Location: platformLocation{
				Type: "PACKAGE",
				Path: filepath.FromSlash("/"),
			},
			Use: cliUse,
		}
		result = append(result, installation)
	} else if filepath.Base(cliDirectory) == "bin" && hasArchiveInstallation(filepath.Dir(cliDirectory)) {
		installation := platformInstallation{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: filepath.Dir(cliDirectory),
			},
			Use: cliUse,
		}
		result = append(result, installation)
	}

	return compactDuplicateInstallations(result), nil
}

func hasArchiveInstallation(dir string) bool {
	if filepath.ToSlash(dir) == "/usr" {
		return false
	}

	return utils.DoesPathExist(filepath.Join(dir, filepath.FromSlash("share/java/confluent-common")))
}

func compactDuplicateInstallations(installations []platformInstallation) []platformInstallation {
	var uniqueInstallations []platformInstallation

	set := types.NewSet[platformLocation]()
	for _, installation := range installations {
		if !set.Contains(installation.Location) {
			set.Add(installation.Location)
			uniqueInstallations = append(uniqueInstallations, installation)
		}
	}

	return uniqueInstallations
}

func choosePluginDir(installation *platformInstallation, prompt form.Prompt, force bool) (string, error) {
	var defaultPluginDir string
	switch installation.Location.Type {
	case "ARCHIVE":
		defaultPluginDir = filepath.Join(installation.Location.Path, "share/confluent-hub-components")
	case "PACKAGE":
		defaultPluginDir = "/usr/share/confluent-hub-components"
	default:
		return "", errors.Errorf(unexpectedInstallationErrorMsg, installation.Location.Type)
	}

	if force {
		output.Printf("Using \"%s\" as the plugin installation directory.\n\n", defaultPluginDir)
		return defaultPluginDir, nil
	}

	f := form.New(form.Field{
		ID:        "confirm",
		Prompt:    fmt.Sprintf(`Do you want to install this plugin into "%s"?`, defaultPluginDir),
		IsYesOrNo: true,
	})
	if err := f.Prompt(prompt); err != nil {
		return "", err
	}
	if f.Responses["confirm"].(bool) {
		output.Print("\n")
		return defaultPluginDir, nil
	}

	f = form.New(form.Field{
		ID:     "directory",
		Prompt: "Specify plugin installation directory. To cancel, press Ctrl-C",
	})
	if err := f.Prompt(prompt); err != nil {
		return "", err
	}

	inputDir := f.Responses["directory"].(string)
	inputDir, err := filepath.Abs(inputDir)
	if err != nil {
		return "", err
	}
	if !utils.DoesPathExist(inputDir) {
		return "", errors.Errorf(invalidDirectoryErrorMsg, inputDir)
	}

	output.Print("\n")
	return inputDir, nil
}

func standardWorkerConfigLocations(installation *platformInstallation) ([]WorkerConfig, error) {
	workerConfigLocations := []string{
		"/etc/kafka/connect-distributed.properties",
		"/etc/kafka/connect-standalone.properties",
		"/etc/schema-registry/connect-avro-distributed.properties",
		"/etc/schema-registry/connect-avro-standalone.properties",
	}
	switch installation.Location.Type {
	case "ARCHIVE":
		var result []WorkerConfig
		for _, workerConfigLocation := range workerConfigLocations {
			workerConfigPath := filepath.Join(installation.Location.Path, filepath.FromSlash(workerConfigLocation))
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
				return nil, errors.Wrapf(err, `failed to read possible $CONFLUENT_CURRENT file "%s"`, confluentCurrentFile)
			}
			confluentCurrentLines := strings.SplitN(string(confluentCurrentContent), "\n", 3)
			if len(confluentCurrentLines) == 1 {
				connectCurrentConfigFile := filepath.Join(confluentCurrentLines[0], "/connect/connect.properties")
				result = append(result, WorkerConfig{Path: connectCurrentConfigFile, Use: "$CONFLUENT_CURRENT"})
			}
		}
		return result, nil
	case "PACKAGE":
		var result []WorkerConfig
		for _, workerConfigLocation := range workerConfigLocations {
			result = append(result, WorkerConfig{Path: filepath.FromSlash(workerConfigLocation), Use: "Standard"})
		}
		return result, nil
	default:
		return nil, errors.New(fmt.Sprintf(unexpectedInstallationErrorMsg, installation.Location.Type))
	}
}

func runningWorkerConfigLocations(searchProcessCmd exec.Command) ([]WorkerConfig, error) {
	re := regexp.MustCompile(workerProcessRegexStr)

	out, err := searchProcessCmd.Output()
	if err != nil {
		return nil, errors.Wrap(err, "failed to run shell command to locate running Connect worker processes")
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

func chooseWorkerConfigs(cmd *cobra.Command, installation *platformInstallation, prompt form.Prompt, force bool) ([]string, error) {
	var workerConfigs []WorkerConfig

	if standardWorkerConfigs, err := standardWorkerConfigLocations(installation); err != nil {
		return nil, errors.Wrap(err, "could not infer possible worker configuration file locations from standard candidates")
	} else {
		for _, workerConfig := range standardWorkerConfigs {
			if utils.DoesPathExist(workerConfig.Path) {
				workerConfigs = append(workerConfigs, workerConfig)
			}
		}
	}

	re := regexp.MustCompile(workerProcessRegexStr)
	commandStr := `ps ax |
					grep -E '` + re.String() + `'|
					grep -v grep;
				test ${PIPESTATUS[0]} -eq 0`
	searchProcessCmd := exec.NewCommand("/bin/bash", "-c", commandStr)

	if runningWorkerConfigs, err := runningWorkerConfigLocations(searchProcessCmd); err != nil {
		return nil, errors.Wrap(err, "could not infer possible worker configuration file locations from running processes")
	} else {
		for _, workerConfig := range runningWorkerConfigs {
			if utils.DoesPathExist(workerConfig.Path) {
				workerConfigs = append(workerConfigs, workerConfig)
			}
		}
	}

	var filteredWorkerConfigs []WorkerConfig
	if len(workerConfigs) == 0 {
		output.Println("No worker configuration files found.")
		return []string{}, nil
	}

	output.Println("Detected the following worker configuration files:")
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

	if force {
		filteredWorkerConfigs = workerConfigs
	} else {
		f := form.New(form.Field{
			ID:        "confirm",
			Prompt:    "Do you want to update all detected config files?",
			IsYesOrNo: true,
		})
		if err := f.Prompt(prompt); err != nil {
			return nil, err
		}
		if f.Responses["confirm"].(bool) {
			filteredWorkerConfigs = workerConfigs
		} else {
			for i, workerConfig := range workerConfigs {
				f := form.New(form.Field{
					ID:        "confirm",
					Prompt:    fmt.Sprintf(`Do you want to update worker configuration file %d?`, i+1),
					IsYesOrNo: true,
				})
				if err := f.Prompt(prompt); err != nil {
					return nil, err
				}
				if f.Responses["confirm"].(bool) {
					filteredWorkerConfigs = append(filteredWorkerConfigs, workerConfig)
				}
			}
		}
	}

	result := make([]string, len(filteredWorkerConfigs))
	for i, workerConfig := range filteredWorkerConfigs {
		result[i] = workerConfig.Path
	}
	return result, nil
}

func updateWorkerConfig(pluginDir string, workerConfigPath string, dryRun bool) error {
	pluginPathProperty := "plugin.path"

	workerConfig, err := properties.LoadFile(workerConfigPath, properties.UTF8)
	if err != nil {
		return errors.Wrapf(err, `failed to parse worker configuration file "%s"`, workerConfigPath)
	}
	pluginPath := workerConfig.GetString(pluginPathProperty, "")
	pluginPathElements := regexp.MustCompile(" *, *").Split(pluginPath, -1)
	for _, pluginPathElement := range pluginPathElements {
		if pluginPathElement == pluginDir {
			output.Printf("This plugin is already in the plugin path for worker configuration file \"%s\".\n", workerConfigPath)
			return nil
		}
	}
	newPluginPath := strings.Join(append(pluginPathElements, pluginDir), ", ")
	if _, _, err = workerConfig.Set(pluginPathProperty, newPluginPath); err != nil {
		return errors.Wrapf(err, `failed to update %s property to "%s" for worker configuration "%s"`, pluginPathProperty, newPluginPath, workerConfigPath)
	}
	fileInfo, err := os.Stat(workerConfigPath)
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	workerConfigFile, err := os.OpenFile(workerConfigPath, os.O_TRUNC|os.O_RDWR, fileInfo.Mode())
	if err != nil {
		return errors.Wrapf(err, `failed to open worker configuration file "%s" before updating with new %s value "%s"`, workerConfigPath, pluginPathProperty, newPluginPath)
	}
	defer workerConfigFile.Close()
	// NOTE: This currently changes the comment spacing and removes empty lines
	if _, err = workerConfig.WriteFormattedComment(workerConfigFile, properties.UTF8); err != nil {
		return errors.Wrapf(err, `failed to update worker configuration file "%s" with new %s value "%s"`, workerConfigPath, pluginPathProperty, newPluginPath)
	}
	return nil
}
